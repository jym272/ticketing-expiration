package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nats-io/nats.go"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"workspace/listeners"

	nt "workspace/nats"
)

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	ctx := nt.GetInstance()

	ctx.AddStreams()
	ctx.ConnectToNats()
	ctx.VerifyStreams()
	ctx.VerifyConsumers()

	go signalListener(signals, ctx.Nc)

	defer func(nc *nats.Conn) {
		err := nc.Drain()
		if err != nil {
			log.Fatalf("Error draining: %v", err)
		}
		fmt.Println("Connection drained.")
		nc.Close()

	}(ctx.Nc)

	//go async.StartServer()
	go nt.Subscribe(nt.OrderCreated, listeners.OrderCreated)

	//mux := http.NewServeMux()
	//
	//mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
	//	type HealthCheckResponse struct {
	//		Status string `json:"status"`
	//	}
	//	response := &HealthCheckResponse{Status: "ok"}
	//	jsonResponse, err := json.Marshal(response)
	//	if err != nil {
	//		http.Error(w, err.Error(), http.StatusInternalServerError)
	//		return
	//	}
	//	w.Header().Set("Content-Type", "application/json")
	//	w.WriteHeader(http.StatusOK)
	//	w.Write(jsonResponse)
	//})
	//
	//// add a handler to the mux for the root endpoint
	//mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	//	fmt.Fprintln(w, "Hello, World!")
	//})
	//
	//// create a new server
	//server := &http.Server{
	//	Addr:    ":8080",
	//	Handler: mux,
	//}
	//
	//// start the server in a goroutine
	//go func() {
	//	err := server.ListenAndServe()
	//	if err != nil {
	//		log.Fatalf("Error starting server: %v", err)
	//	}
	//}()

	// wait for user input to exit the program
	//fmt.Println("Press Enter to exit.")
	//fmt.Scanln()

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return c.HTML(http.StatusOK, "Hello, Docker! <3")
	})

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, struct {
			Status string `json:"status"`
		}{Status: "OK"})
	})

	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	e.Logger.Fatal(e.Start(":" + httpPort))
}

func signalListener(signals chan os.Signal, nc *nats.Conn) {
	<-signals
	fmt.Println("Received signal.")
	err := nc.Drain()

	if err != nil {
		fmt.Println("Error draining.", err)
	}

	fmt.Println("Connection drained.")

	os.Exit(0)
}
