# Load extensions
load('ext://restart_process', 'docker_build_with_restart')

### K8s Configuration ###
# Apply namespace, configmap, and secrets
k8s_yaml('./infra/development/k8s/app-config.yaml')

### API Services ###
api_compile_cmd = 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../../build/api-services ./cmd'
if os.name == 'nt':
  api_compile_cmd = 'powershell -ExecutionPolicy Bypass -File build.ps1'

local_resource(
  'api-services-compile',
  api_compile_cmd,
  deps=['./services/api-services'],
  labels=["compiles"],
  dir='./services/api-services'
)

docker_build_with_restart(
  'uptime/api-services:dev',
  '.',
  entrypoint=['sh', '-c', 'touch /tmp/.restart-proc && /app/api-service'],
  dockerfile='./infra/development/docker/api-service.Dockerfile',
  only=[
    './services/api-services',
  ],
  live_update=[
    sync('./services/api-services', '/app/services/api-services'),
  ],
  restart_file='/app/.restart-proc',
)

k8s_yaml('./infra/development/k8s/api-service-deployment.yaml')
k8s_resource('api-services', 
             port_forwards=['8082:8082'],
             resource_deps=['api-services-compile'], 
             labels=["services"])

### Web Frontend ###
# For development, we'll use live reload with Next.js dev server
local_resource(
  'web-dev',
  'npm run dev',
  deps=['./web'],
  serve_cmd='npm run dev',
  dir='./web',
  resource_deps=['api-services'],
  labels=["web"]
)

# Alternative: Build and deploy web as container (uncomment if needed)
# docker_build(
#   'uptime/web:dev',
#   '.',
#   dockerfile='./infra/development/docker/web.Dockerfile',
#   only=[
#     './web',
#   ],
#   live_update=[
#     sync('./web', '/app'),
#     run('npm install', trigger=['./web/package.json', './web/package-lock.json']),
#   ],
# )
# 
# k8s_yaml('./infra/development/k8s/web-deployment.yaml')
# k8s_resource('web', 
#              port_forwards=['3000:3000'],
#              labels=["web"])

### Development Helpers ###
# Watch for changes in shared modules
local_resource(
  'shared-watcher',
  'echo "Watching shared modules for changes..."',
  deps=['./shared'],
  labels=["watchers"]
)

# Health check endpoint for all services
local_resource(
  'health-check',
  'echo "Health check endpoints:" && echo "API Services: http://localhost:8082/health" && echo "Web: http://localhost:3000/api/health"',
  labels=["info"]
)
