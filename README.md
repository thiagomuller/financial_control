Run with:
```bash
podman-compose -f /home/$USER/financial_control/docker-compose.yml build && \
podman ps -a --filter "name=my-finances-ui" --format "{{.Names}}" | xargs -r podman stop && \
podman ps -a --filter "name=my-finances-api" --format "{{.Names}}" | xargs -r podman stop && \
podman ps -a --filter "name=my-finances-db" --format "{{.Names}}" | xargs -r podman stop && \
podman ps -a --filter "name=my-finances-ui" --format "{{.Names}}" | xargs -r podman rm && \
podman ps -a --filter "name=my-finances-api" --format "{{.Names}}" | xargs -r podman rm && \
podman ps -a --filter "name=my-finances-db" --format "{{.Names}}" | xargs -r podman rm && \
podman-compose -f /home/$USER/financial_control/docker-compose.yml up -d
```

Stop with:
```bash
podman-compose -f /home/$USER/financial_control/docker-compose.yml down && \
podman ps -a --filter "name=my-finances-ui" --format "{{.Names}}" | xargs -r podman stop && \
podman ps -a --filter "name=my-finances-api" --format "{{.Names}}" | xargs -r podman stop && \
podman ps -a --filter "name=my-finances-db" --format "{{.Names}}" | xargs -r podman stop && \
podman ps -a --filter "name=my-finances-ui" --format "{{.Names}}" | xargs -r podman rm && \
podman ps -a --filter "name=my-finances-api" --format "{{.Names}}" | xargs -r podman rm && \
podman ps -a --filter "name=my-finances-db" --format "{{.Names}}" | xargs -r podman rm 
```

> Replace podman with docker, in case you're using docker as the container engine
> Replace /home/$USER/financial_control/docker-compose.yml with the path you've cloned this repo into
