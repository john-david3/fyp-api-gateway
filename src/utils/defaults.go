package utils

var (
	GatewayConfigDirName  = "/etc/config/"
	GatewayConfigFileName = "gateway.yaml"

	NGINXDirName          = "/etc/nginx/"
	NGINXTemplateDirName  = "/etc/nginx/"
	NGINXConfigFileName   = "nginx.conf"
	NGINXTemplateFileName = "nginx.conf.tmpl"
	NGINXUserDirName      = "/etc/nginx/users/"

	DefaultConfigContent = "#\n# Configuration file for API Gateway\n#\n\nconnections:\n  routes:\n    - path: /products\n      url: http://services:9001\n      rate-limit:\n        zone: 10\n        rate: 5\n      auth: false\n      zone-name: root\n\n    - path: /orders\n      url: http://services:9002\n      rate-limit:\n        zone: 10\n        rate: 5\n      auth: false\n      zone-name: root\n\n    - path: /protected\n      url: http://services:9003\n      rate-limit:\n        zone: 10\n        rate: 5\n      auth: true\n      zone-name: root\n\n    - path: /external-weather\n      url: https://api.open-meteo.com/v1/forecast?latitude=51.898&longitude=-8.4706&hourly=temperature_2m/\n      rate-limit:\n        zone: 10\n        rate: 5\n      auth: false\n      zone-name: root"
)
