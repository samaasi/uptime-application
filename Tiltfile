# Load the restart_process extension
load('ext://restart_process', 'docker_build_with_restart')

### K8s Config ###
# Uncomment to use secrets
# k8s_yaml('./infra/development/k8s/secrets.yaml')

k8s_yaml('./infra/development/k8s/app-config.yaml')
### End of K8s Config ###

### API Gateway ###
gateway_compile_cmd = 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/api-gateway ./services/api-gateway/cmd'
if os.name == 'nt':
  gateway_compile_cmd = 'infra\\development\\docker\\api-gateway-build.bat'

local_resource(
  'api-gateway-compile',
  gateway_compile_cmd,
  deps=['./services/api-gateway', './shared'], labels="compiles")


docker_build_with_restart(
  'uptime-application/services/api-gateway',
  '.',
  entrypoint=['/app/build/api-gateway'],
  dockerfile='./infra/development/docker/api-gateway.Dockerfile',
  only=[
    './build/api-gateway',
    './shared',
  ],
  live_update=[
    sync('./build', '/app/build'),
    sync('./shared', '/app/shared'),
  ],
)

k8s_yaml('./infra/development/k8s/api-gateway-deployment.yaml')
k8s_resource('api-gateway', port_forwards=8081,
             resource_deps=['api-gateway-compile'], labels="services")

### Web service ###
docker_build(
  'uptime-application/web',
  '.',
  dockerfile='./infra/development/docker/web.Dockerfile',
)

k8s_yaml('./infra/development/k8s/web-deployment.yaml')
k8s_resource('web', port_forwards=3000, labels="web_service")
### End of Web Service ###

### API Service ###
api_compile_cmd = 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/api-service ./services/api-service/cmd'
if os.name == 'nt':
  api_compile_cmd = 'infra\\development\\docker\\api-service-build.bat'

local_resource(
  'api-service-compile',
  api_compile_cmd,
  deps=['./services/api-service', './shared'], labels="compiles")

docker_build_with_restart(
  'uptime-application/services/api-service',
  '.',
  entrypoint=['/app/build/api-service'],
  dockerfile='./infra/development/docker/api-service.Dockerfile',
  only=[
    './build/api-service',
    './shared',
  ],
  live_update=[
    sync('./build', '/app/build'),
    sync('./shared', '/app/shared'),
  ],
)

k8s_yaml('./infra/development/k8s/api-service-deployment.yaml')
k8s_resource('api-service', port_forwards=8082,
             resource_deps=['api-service-compile'], labels="services")
