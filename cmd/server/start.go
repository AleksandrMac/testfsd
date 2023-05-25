package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/AleksandrMac/testfsd/docs"
	"github.com/AleksandrMac/testfsd/internal/config"
	"github.com/AleksandrMac/testfsd/internal/connect"
	"github.com/AleksandrMac/testfsd/internal/log"
	internalMetric "github.com/AleksandrMac/testfsd/internal/metric"
	internalTrace "github.com/AleksandrMac/testfsd/internal/trace"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/spf13/cobra"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

var (
	configPath string
	startCmd   = &cobra.Command{
		Use:   "start",
		Short: "start server",
		Long:  `start server, default port is 5000`,
		Run:   startServer,
	}
	enablePprof bool
)

func init() {
	cobra.OnInitialize(initConfig, initLogger)
	rootCmd.AddCommand(startCmd)
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "config file (default is $PWD/config/default.yaml)")
	startCmd.PersistentFlags().Int("port", 5000, "Port to run Application server on")
	startCmd.PersistentFlags().BoolVarP(&enablePprof, "pprof", "p", false, "enable pprof mode (default: false)")
	config.Viper().BindPFlag("port", startCmd.PersistentFlags().Lookup("port"))
}

func initConfig() {
	defer log.Sync()
	if len(configPath) != 0 {
		config.Viper().SetConfigFile(configPath)
	} else {
		config.Viper().AddConfigPath("./config")
		config.Viper().SetConfigName("default")
	}
	config.Viper().SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.Viper().AutomaticEnv()
	if err := config.Viper().ReadInConfig(); err != nil {
		log.Fatal(
			fmt.Sprintf("Load config from file [%s]: %v", config.Viper().ConfigFileUsed(), err))
	}
	config.Parse()
}

func initLogger() {
	log.ResetDefault(log.New(os.Stderr, config.Default.Otel.Log,
		zap.AddCaller(), zap.AddCallerSkip(1)))
}

func startServer(cmd *cobra.Command, agrs []string) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	observabilityCloser := initObservability(ctx)
	defer observabilityCloser(ctx)

	log.Info(fmt.Sprintf("Start http-server on %s:%d", config.Default.Server.Host, config.Default.Server.Port))
	setupDoc()
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RequestLogger(log.Default()))
	r.Use(middleware.RedirectSlashes)
	r.Use(middleware.Recoverer)

	if enablePprof {
		r.Mount("/debug", middleware.Profiler())
	}

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {})
	r.Get("/doc/*", httpSwagger.WrapHandler)

	connect.RegisterRouterConnect(r)

	http.ListenAndServe(fmt.Sprintf("%s:%d", config.Default.Server.Host, config.Default.Server.Port), r)
}

func setupDoc() {
	// programmatically set swagger info
	docs.SwaggerInfo.Title = "Tinkoff"
	docs.SwaggerInfo.Description = "Адаптер к тинькофф банку"
	docs.SwaggerInfo.Version = Version
	docs.SwaggerInfo.Host = fmt.Sprintf("%s:%d", config.Default.Server.Host, config.Default.Server.Port)
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
}

func initObservability(ctx context.Context) (close func(context.Context)) {
	closedFns := []func(context.Context){}

	// tracer initialization
	log.Info("Start trace provider")
	shutdownTrace, err := internalTrace.InitTraceProvider(ctx, config.Default.Metadata.ServiceName, Version, config.Default.Otel.Trace)
	if err != nil {
		if err != internalTrace.ErrUndefindedTraceProto {
			log.Fatal(err.Error())
		}
		log.Info(err.Error())
	}

	if shutdownTrace != nil {
		closedFns = append(closedFns, func(ctx context.Context) {
			if err := shutdownTrace(ctx); err != nil {
				log.Fatal("failed to shutdown TracerProvider: " + err.Error())
			}
		})
	}

	// meter initialization
	log.Info("Start metric provider")
	shutdownMeter, err := internalMetric.InitMeterProvider(ctx, config.Default.Metadata.ServiceName, Version, config.Default.Otel.Metric)
	if err != nil {
		if err != internalMetric.ErrUndefindedMetricProto {
			log.Fatal(err.Error())
		}
		log.Info(err.Error())
	}

	if shutdownMeter != nil {
		closedFns = append(closedFns, func(ctx context.Context) {
			if err := shutdownMeter(ctx); err != nil {
				log.Fatal("failed to shutdown MeterProvider: " + err.Error())
			}
		})
	}

	return func(ctx context.Context) {
		for i := len(closedFns) - 1; i > 0; i-- {
			closedFns[i](ctx)
		}
	}
}
