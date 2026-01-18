package ingredient

import (
	"context"
	_ "embed"
	"errors"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/dtm-labs/rockscache"
	jsoniter "github.com/json-iterator/go"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
)

//go:embed system.txt
var PromptSystem string

//go:embed user.txt
var PromptUser string

var (
	ErrorMissingKey      = errors.New("key is required")
	ErrorMissingRedisDsn = errors.New("redis dsn is required")
)

const (
	DefaultCacheExpireMinutes = 60 * 24 * 7
	CacheRedisPrefix          = "bytes:ingredient:cache:"

	DefaultEmbeddingCacheExpireMinutes = 60 * 24 * 7 * 30
	CacheEmbeddingRedisPrefix          = "bytes:embedding:cache:"
)

type Config struct {
	Endpoint     string `koanf:"endpoint"`
	Key          string `koanf:"key"`
	Model        string `koanf:"model"`
	RedisDsn     string `koanf:"redis_dsn"`
	CacheMinutes int    `koanf:"cache_minutes"`
}

type Analysis interface {
	Analyze(dishes []string) ([]string, error)
	Embedding(inputs []string, trim bool) (map[string]string, error)
}

type analysisImpl struct {
	Config

	AiClient *openai.Client
	rcClient *rockscache.Client
}

type Option func(analysis *analysisImpl) error

func WithConfig(config Config) Option {
	return func(analysis *analysisImpl) error {
		analysis.Config = config
		return nil
	}
}

func WithConfigFile(filename string) Option {
	return func(analysis *analysisImpl) error {
		var k = koanf.New(".")
		if err := k.Load(file.Provider(filename), yaml.Parser()); err != nil {
			return err
		}

		var config Config
		if err := k.Unmarshal("ingredient", &config); err != nil {
			return err
		}

		return nil
	}
}

func New(options ...Option) (Analysis, error) {
	instance := &analysisImpl{}
	for _, option := range options {
		if err := option(instance); err != nil {
			return nil, err
		}
	}

	if instance.Endpoint == "" {
		envEndpoint, envEndpointExist := os.LookupEnv("INGREDIENT_ENDPOINT")
		if envEndpointExist {
			instance.Endpoint = envEndpoint
		}

		if instance.Endpoint == "" {
			instance.Endpoint = "https://api.openai.com/v1"
		}
	}

	if instance.Key == "" {
		envKey, envKeyExist := os.LookupEnv("INGREDIENT_KEY")
		if envKeyExist {
			instance.Key = envKey
		}

		if instance.Key == "" {
			log.Printf("[ingredient] key is required")
			return nil, ErrorMissingKey
		}
	}

	if instance.Model == "" {
		envModel, envModelExist := os.LookupEnv("INGREDIENT_MODEL")
		if envModelExist {
			instance.Model = envModel
		}

		if instance.Model == "" {
			instance.Model = openai.GPT4o
		}
	}

	if instance.RedisDsn == "" {
		envRedisDsn, envRedisDsnExist := os.LookupEnv("INGREDIENT_REDIS_DSN")
		if envRedisDsnExist {
			instance.RedisDsn = envRedisDsn
		}

		if instance.RedisDsn == "" {
			log.Printf("[ingredient] redis dsn is required")
			return nil, ErrorMissingRedisDsn
		}
	}

	if instance.CacheMinutes == 0 {
		envCacheMinutes, envCacheMinutesExist := os.LookupEnv("INGREDIENT_CACHE_MINUTES")
		if envCacheMinutesExist {
			instance.CacheMinutes = cast.ToInt(envCacheMinutes)
		}

		if instance.CacheMinutes == 0 {
			instance.CacheMinutes = DefaultCacheExpireMinutes
		}
	}

	redisOptions, err := redis.ParseURL(instance.RedisDsn)
	if err != nil {
		return nil, err
	}

	redisClient := redis.NewClient(redisOptions)
	if redisClient.Ping(context.Background()).Err() != nil {
		log.Printf("[ingredient] failed to connect to redis: %v", err)
		return nil, err
	}
	instance.rcClient = rockscache.NewClient(redisClient, rockscache.NewDefaultOptions())

	aiConfig := openai.DefaultConfig(instance.Key)
	aiConfig.BaseURL = instance.Endpoint
	instance.AiClient = openai.NewClientWithConfig(aiConfig)

	return instance, nil
}

func (a *analysisImpl) Analyze(dishes []string) ([]string, error) {
	var allIngredients []string

	for _, dish := range dishes {
		rawIngredients, err := a.rcClient.Fetch(
			CacheRedisPrefix+dish,
			time.Duration(a.CacheMinutes)*time.Minute,
			func() (string, error) {
				rawJson, err := a.doAnalysis(dish)
				if err != nil {
					return "", err
				}

				return rawJson, nil
			},
		)

		if err != nil {
			log.Printf("[ingredient] failed to fetch ingredients: %v", err)
			return nil, err
		}

		jsonIngredients := gjson.Get(rawIngredients, "ingredients")
		if jsonIngredients.Exists() && jsonIngredients.IsArray() {
			lo.ForEach(jsonIngredients.Array(), func(value gjson.Result, _ int) {
				elem := value.String()
				if elem != "" {
					allIngredients = append(allIngredients, elem)
				}
			})
		}
	}

	return lo.Uniq(allIngredients), nil
}

func (a *analysisImpl) Embedding(inputs []string, trim bool) (map[string]string, error) {
	embeddingMap := make(map[string]string)

	if len(inputs) == 0 {
		return embeddingMap, nil
	}

	inputs = lo.Uniq(inputs)
	keys := lo.Map(inputs, func(value string, _ int) string {
		return CacheEmbeddingRedisPrefix + value
	})

	strEmbeddings, err := a.rcClient.FetchBatch(
		keys,
		time.Duration(DefaultEmbeddingCacheExpireMinutes)*time.Minute,
		func(missingIdxs []int) (map[int]string, error) {
			missingEmbeddingMap := make(map[int]string)

			missingInputs := lo.Map(
				missingIdxs,
				func(inputIdx int, _ int) string {
					if trim {
						return trimMenuName(inputs[inputIdx])
					}

					return inputs[inputIdx]
				},
			)

			resp, err := a.AiClient.CreateEmbeddings(
				context.Background(),
				openai.EmbeddingRequestStrings{
					Model: openai.SmallEmbedding3,
					Input: missingInputs,
				},
			)

			if err != nil {
				log.Printf("[ingredient] failed to create embedding: %v", err)
				return nil, err
			}

			for _, data := range resp.Data {
				if data.Index < 0 || data.Index >= len(missingIdxs) {
					continue
				}

				sourceIdx := missingIdxs[data.Index]
				if strEmbedding, err := jsoniter.MarshalToString(data.Embedding); err != nil {
					log.Printf("[ingredient] failed to marshal embedding: %v", err)
					continue
				} else {
					missingEmbeddingMap[sourceIdx] = strEmbedding
				}
			}

			return missingEmbeddingMap, nil
		},
	)

	if err != nil {
		log.Printf("[ingredient] failed to fetch embeddings: %v", err)
		return nil, err
	}

	for idx, embedding := range strEmbeddings {
		if embedding == "" {
			continue
		}

		embeddingMap[inputs[idx]] = embedding
	}

	return embeddingMap, nil
}

func (a *analysisImpl) doAnalysis(dish string) (string, error) {
	if dish == "" {
		return "", nil
	}

	resp, err := a.AiClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: a.Model,
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: PromptSystem,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: PromptUser + dish,
				},
			},
		},
	)

	if err != nil {
		log.Printf("[ingredient] failed to create chat completion: %v", err)
		return "", err
	}

	rawIngredients := resp.Choices[0].Message.Content
	jsonIngredients := gjson.Get(rawIngredients, "ingredients")
	if jsonIngredients.Exists() && jsonIngredients.IsArray() {
		return rawIngredients, nil
	}

	return "", nil
}

func trimMenuName(name string) string {
	menuNameRe := regexp.MustCompile(`\(.*\)`)
	return strings.TrimSpace(menuNameRe.ReplaceAllString(name, ""))
}
