# Trail Simulator
The purpose of this simulator is to explain Trail architecture.

## Intallation
If `/go/src/` directory does't exist, run `mkdir ~/go/src/`.

Please install this repository under `/go/src/`.

## Run simulator
If you do not use Docker, install golang and run `golang run simulator/src/main.go`.

If you use Docker, install Docker. Then run `make start` and `golang run src/main.go`.

The simulation parameters are in `simulator/src/setting/setting.go`.

## Notes
This simulator has not been developed for experiments.
If you set a large value for the parameter, the simulator wastes computer resources and the simulation takes a long time.

**Set NumberOfClient to 500 or less and EndBlockHeight to 200 or less.**



