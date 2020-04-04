# NukiGoBridge - A nuki bridge implemented in golang

## Sponsoring

- Buy me a coffee: https://www.paypal.me/JochenScheib
- Buy me a present: https://www.amazon.de/hz/wishlist/ls/2ORJGSSLEQDED?ref_=wl_share

## Usage

 ### Docker

 #### Requirements

 To access the hardware interface for bluetooth connections the following parameters must be set
`--cap-add=SYS_ADMIN --cap-add=NET_ADMIN --net=host`

For more information why that is needed see: https://lore.kernel.org/patchwork/patch/820786/
 

 #### Volumes

 Volume | Description
 -------|------------
 /config | Persistant storage for configuration

 #### Environment variables

 Variable | Default | Description
 ---------|---------|------------
 NUKI_TOKEN | generated during start | Used to authenticate api calls, if not set token will be generated on each restart
 NUKI_CONFIGPATH | /config | Used to store the configuration file, including paired locks
 PORT | 8080 | HTTP server port for api

 #### Example Usage

 ```
 docker run --name nukibridge -e NUKI_TOKEN=secret1234 -v /mnt/storage/nukibridge-config:/config --rm --cap-add=SYS_ADMIN --cap-add=NET_ADMIN --net=host bashgroup/nukigobridge
 ```
- `--name nukibridge`: Giving the container a name
- `-e NUKI_TOKEN=secret1234`: Setting the token to access the api
- `-v /mnt/storage/nukibridge-config:/config`: Mount persistant storage as volume
- `--rm`: Delete container when stopped
- `--cap-add=SYS_ADMIN --cap-add=NET_ADMIN --net=host`: Needed for bluetooth
- `bashgroup/nukigobridge:latest`: The image

### API

The bridge provides an api vi http. It is splitted into two parts

- Replication of the official api
- Extended restful api

For details see *assets/doc*

The api documentation can be viewed and tested after the bridge runs under `http://<ip>:8080/doc` using swagger ui.

### ToDo

- [x] Automated builds
- [ ] Better api access using jwt token
- [ ] Extend api by more functionality
- [ ] Support nuki opener (needs sponsoring)

## License

See *License*
