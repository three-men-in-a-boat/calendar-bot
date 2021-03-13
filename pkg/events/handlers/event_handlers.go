package handlers

import (
	"github.com/labstack/echo"
	"go.uber.org/zap"
)

type EventHandlers struct {
	//albumLogic albums.AlbumUseCase
	logger     *zap.SugaredLogger
}

func NewEventHandlers(log *zap.SugaredLogger, albumRealisation albums.AlbumUseCase) AlbumDeliveryRealisation {
	return AlbumDeliveryRealisation{albumLogic: albumRealisation, logger: log}
}

func (event AlbumDeliveryRealisation) InitHandlers(server *echo.Echo) {

	server.POST("api/v1/album", albumD.CreateAlbum)

	server.GET("api/v1/albums", albumD.GetAlbums)

}
