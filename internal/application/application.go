package application

import (
	"errors"
	"fmt"
	"io"

	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/handler"
	"github.com/alex-storchak/shortener/internal/middleware"
	"github.com/alex-storchak/shortener/internal/repository"
	"github.com/alex-storchak/shortener/internal/repository/factory"
	"github.com/alex-storchak/shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/teris-io/shortid"
	"go.uber.org/zap"
)

type closer struct {
	name   string
	closer io.Closer
}

type closers []closer

type App struct {
	cfg     *config.Config
	logger  *zap.Logger
	router  *chi.Mux
	closers closers
}

func NewApp(cfg *config.Config, l *zap.Logger) (*App, error) {
	app := &App{
		cfg:    cfg,
		logger: l,
	}

	sf, err := factory.NewStorageFactory(cfg, l)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage factory: %w", err)
	}
	shortener, err := app.initShortener(sf)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize shortener: %w", err)
	}
	if err := app.initRouter(shortener, sf); err != nil {
		return nil, fmt.Errorf("failed to initialize router: %w", err)
	}
	return app, nil
}

func (a *App) Close() error {
	var errs []error
	for i := len(a.closers) - 1; i >= 0; i-- {
		if err := a.closers[i].closer.Close(); err != nil {
			err = fmt.Errorf("failed to close `%s`: %w", a.closers[i].name, err)
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (a *App) Run() error {
	return handler.Serve(&a.cfg.Handler, a.router)
}

func (a *App) initURLStorage(sf factory.IStorageFactory) (repository.URLStorage, error) {
	storage, err := sf.MakeURLStorage()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize url storage: %w", err)
	}
	a.closers = append(a.closers, closer{"storage", storage})
	return storage, nil
}

func (a *App) initShortIDGenerator() (*service.ShortIDGenerator, error) {
	generator, err := shortid.New(1, shortid.DefaultABC, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate shortid generator: %w", err)
	}
	return service.NewShortIDGenerator(generator), nil
}

func (a *App) initShortener(sf factory.IStorageFactory) (*service.Shortener, error) {
	gen, err := a.initShortIDGenerator()
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate shortid generator: %w", err)
	}
	storage, err := a.initURLStorage(sf)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize url storage: %w", err)
	}
	a.logger.Info("shortener initialized")
	return service.NewShortener(gen, storage, a.logger), nil
}

func (a *App) initHandlers(shortener service.IShortenerService) *handler.Handlers {
	mainPageService := service.NewMainPageService(a.cfg.Handler.BaseURL, shortener, a.logger)
	mainPageHandler := handler.NewMainPageHandler(mainPageService, a.logger)

	shortURLService := service.NewShortURLService(shortener, a.logger)
	shortURLHandler := handler.NewShortURLHandler(shortURLService, a.logger)

	jsonDecoder := service.JSONRequestDecoder{}
	apiShortenService := service.NewAPIShortenService(a.cfg.Handler.BaseURL, shortener, jsonDecoder, a.logger)
	jsonEncoder := service.JSONEncoder{}
	apiShortenHandler := handler.NewAPIShortenHandler(apiShortenService, jsonEncoder, a.logger)

	batchDecoder := service.JSONBatchRequestDecoder{}
	apiShortenBatchService := service.NewAPIShortenBatchService(a.cfg.Handler.BaseURL, shortener, batchDecoder, a.logger)
	apiShortenBatchHandler := handler.NewAPIShortenBatchHandler(apiShortenBatchService, jsonEncoder, a.logger)

	apiUserURLsService := service.NewAPIUserURLsService(a.cfg.Handler.BaseURL, shortener, batchDecoder, a.logger)
	apiUserURLsHandler := handler.NewAPIUserURLsHandler(apiUserURLsService, jsonEncoder, a.logger)

	pingService := service.NewPingService(shortener, a.logger)
	pingHandler := handler.NewPingHandler(pingService, a.logger)

	return &handler.Handlers{
		mainPageHandler,
		shortURLHandler,
		apiShortenHandler,
		apiShortenBatchHandler,
		apiUserURLsHandler,
		pingHandler,
	}
}

func (a *App) initMiddlewares(sf factory.IStorageFactory) (*handler.Middlewares, error) {
	us, err := sf.MakeUserStorage()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize user storage: %w", err)
	}
	as := service.NewAuthService(a.logger, us, &a.cfg.Middleware)
	um := repository.NewUserManager(a.logger, us)
	authMWService := service.NewAuthMiddlewareService(as, um, &a.cfg.Middleware)
	return &handler.Middlewares{
		middleware.RequestLogger(a.logger),
		middleware.AuthMiddleware(a.logger, authMWService, &a.cfg.Middleware),
		middleware.GzipMiddleware(a.logger),
	}, nil
}

func (a *App) initRouter(shortener service.IShortenerService, sf factory.IStorageFactory) error {
	handlers := a.initHandlers(shortener)
	middlewares, err := a.initMiddlewares(sf)
	if err != nil {
		return fmt.Errorf("failed to initialize middlewares: %w", err)
	}
	a.router = handler.NewRouter(handlers, middlewares)
	a.logger.Info("router initialized")
	return nil
}
