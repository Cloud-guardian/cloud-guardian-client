Easy client:

- Rename cloud-guardian to cloud-guardian-ez-<apikey>

Build for environments:

```
# production:
make clean release

# staging:
API_URL="https://api.cloud-guardian.net/cloudguardian-api-staging/v1/" make clean release

# local development:
API_URL="http://localhost:8080/cloudguardian-api/v1/" make clean release

# local development inside docker container:
API_URL="http://host.docker.internal:8080/cloudguardian-api/v1/" make clean release
```


Start the docker container:
```
make run-docker-rocky-9
```

Then inside the container:
```
/client/linux_amd64/cloud-guardian --api-key $API_KEY --register

/client/linux_amd64/cloud-guardian --api-key $API_KEY
```
Example build targets:

```
make run-docker-almalinux-9.3
make run-docker-rockylinux-9.3
make run-docker-rockylinux-9.5
make run-docker-rockylinux-9.5.20241118
make run-docker-ubuntu-jammy-20240808
```
