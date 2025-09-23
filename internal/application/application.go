package application

import (
	"errors"
	"fmt"
	"io"

	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/handler"
	"github.com/alex-storchak/shortener/internal/logger"
	"github.com/alex-storchak/shortener/internal/middleware"
	"github.com/alex-storchak/shortener/internal/repository"
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

func NewApp() (*App, error) {
	app := &App{}

	if err := app.initConfig(); err != nil {
		return nil, fmt.Errorf("failed to initialize config: %w", err)
	}
	if err := app.initLogger(); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	shortener, err := app.initShortener()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize shortener: %w", err)
	}

	app.initRouter(shortener)
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

	if err := a.logger.Sync(); err != nil {
		err = fmt.Errorf("failed to sync logger: %w", err)
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (a *App) Run() error {
	return handler.Serve(&a.cfg.Handler, a.router)
}

func (a *App) initConfig() error {
	cfg, err := config.ParseConfig()
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}
	a.cfg = cfg
	return nil
}

func (a *App) initLogger() error {
	zl, err := logger.NewLogger(&a.cfg.Logger)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	a.logger = zl
	zl.Info("logger initialized")
	return nil
}

func (a *App) initURLStorage() (repository.URLStorage, error) {
	sf := repository.NewStorageFactory(a.cfg, a.logger)
	storage, err := sf.Produce()
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

func (a *App) initShortener() (*service.Shortener, error) {
	gen, err := a.initShortIDGenerator()
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate shortid generator: %w", err)
	}
	storage, err := a.initURLStorage()
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

	pingService := service.NewPingService(shortener, a.logger)
	pingHandler := handler.NewPingHandler(pingService, a.logger)

	return &handler.Handlers{
		mainPageHandler,
		shortURLHandler,
		apiShortenHandler,
		apiShortenBatchHandler,
		pingHandler,
	}
}

func (a *App) initMiddlewares() *handler.Middlewares {
	return &handler.Middlewares{
		middleware.RequestLogger(a.logger),
		middleware.GzipMiddleware(a.logger),
	}
}

func (a *App) initRouter(shortener service.IShortenerService) {
	handlers := a.initHandlers(shortener)
	middlewares := a.initMiddlewares()
	a.router = handler.NewRouter(handlers, middlewares)
	a.logger.Info("router initialized")
}
