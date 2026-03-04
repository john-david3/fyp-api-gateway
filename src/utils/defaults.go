package utils

var (
	GatewayConfigDirName  = "/etc/config/"
	GatewayConfigFileName = "gateway.yaml"

	NGINXDirName          = "/etc/nginx/"
	NGINXTemplateDirName  = "/etc/nginx/"
	NGINXConfigFileName   = "nginx.conf"
	NGINXTemplateFileName = "nginx.conf.tmpl"
	NGINXUserDirName      = "/etc/nginx/users/"

	DefaultConfigContent = "#\n# Configuration file for API Gateway\n#\n\nconnections:\n  - host: localhost\n    port: 8080\n    routes:\n      - path: /products\n        upstream:\n          name: product_service\n          port: 9001\n        rate-limit:\n          zone: 10\n          rate: 5\n        auth: false\n\n      - path: /orders\n        upstream:\n          name: order_service\n          port: 9002\n        rate-limit:\n          zone: 10\n          rate: 5\n        auth: false\n\n      - path: /protected\n        upstream:\n          name: protected_service\n          port: 9003\n        rate-limit:\n          zone: 10\n          rate: 5\n        auth: true"
)
