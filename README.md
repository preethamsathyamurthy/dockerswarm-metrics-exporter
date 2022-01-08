# dockerswarm-metrics-exporter
A lightweight node and docker swarm metrics exporter written in Go Echo framework

Can be run by
```bash
go run server.go
```

Then via curl 
```bash
curl -X GET http://localhost:1323
```

## Folder Structure

* server.go (starting function)
*  framework (package contains all framework functions)
   * logger.go (contains Struct for passing Loggin function across packages)  
* core (package Contains all core functions)
  *  initialize.go (initializes common logger for core package)
  *  dockerFunctions.go (Contains all Docker related functions)
  *  systemFunctions.go (Contains all System related functions)
*  output (folder contains example output of metrics)  


### Echo Framework General Notes:
Using the same logging function across packages is not supported by defaut in Echo framework

To support this, can be done by adding the below as part of any new project:
* **logger.go** in framework package (add new functions to this wrapper if needed)
* **initialize.go** as part of any new packages to add (replace the package name in the top of the file)
* In the main package, 
  * ```bash
  	var writeLogs framework.EchoLogger
	writeLogs.Initialize(e) 
    ```
    Here e is of type *echo.Echo
 * ```bash
   PACKAGE_NAME.initialize(writeLogs)
   ```
   Replace the PACKAGE_NAME with the right package


