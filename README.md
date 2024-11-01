# Form3 Technical Exercise

## Requirements
* Go >= 1.23

## Info

* The project layout follows https://github.com/golang-standards/project-layout.
    * `cmd/simulator` : entrypoint of the application.
    * `internal/app/simulator` : business logic and infrastructure independent components.
    * `internal/infra` : infrastructure dependent components.
* Application reads configuration from environment variables using `kelseyhightower/envconfig`.
* `stretchr/testify` is used for testing utilities.

## Running application

To run application, run the following command. It will start application at `localhost:11111`.

```shell
make run
```

or

```shell
go run ./cmd/simulator/main.go
```

You can override configuration with the following environment variables.

```
KEY                                     TYPE             DEFAULT  
APP_SERVER_PORT                         Integer          11111    
APP_SERVER_HOST                         String           localhost
APP_SERVER_GRACEFUL_SHUTDOWN_TIMEOUT    Duration         3s       
APP_INIT_DEBUG                          True or False             
APP_DUMMY_MIN_AMOUNT_TO_WAIT            Integer          100      
APP_DUMMY_MAX_AMOUNT_TO_WAIT            Integer          10000    
```

## Running Tests

To run tests, run the following command.

```shell
make test
```