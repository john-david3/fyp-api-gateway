package utils

var (
	NGINXDirName         = "/etc/nginx/"
	NGINXTemplateDirName = "/etc/nginx/"
	NGINXConfigFileName  = "nginx.conf"
	NGINXZoneFileName    = "zone.conf"

	NGINXUserDirName          = "/etc/nginx/users/"
	NGINXTemplateFileName     = "nginx.conf.tmpl"
	NGINXZoneTemplateFileName = "zone.conf.tmpl"

	DefaultConfigContent = "#\n# Configuration file for API Gateway\n#\n\n#connections:\n#  routes:\n#    - path: \n#      url:\n#      rate-limit:\n#        zone: \n#        rate:\n#      auth:"
)
