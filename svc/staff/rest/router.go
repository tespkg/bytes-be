package rest

import (
	"github.com/gin-gonic/gin"
	cors "github.com/rs/cors/wrapper/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/tespkg/bytes-be/svc/staff/middle"
)

func (s *Server) ginRouter() {
	engine := s.engin
	engine.Use(cors.New(s.corsConfig))
	engine.Use(gin.Logger(), gin.Recovery())
	engine.RedirectTrailingSlash = false

	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.URL("doc.json")))

	engine.Use(middle.WithTimezone())

	v1 := engine.Group("/api/v1")

	{
		s.routerCommon(v1.Group("/common"))
	}
	{
		s.routerCustomer(v1.Group("/customer"))
	}
}

func (s *Server) routerCommon(group *gin.RouterGroup, mws ...gin.HandlerFunc) {

}

func (s *Server) routerCustomer(group *gin.RouterGroup, mws ...gin.HandlerFunc) {

	group.Use(middle.WithToken(s.db))
	group.Use(middle.WithUserInfo(s.db))
}
