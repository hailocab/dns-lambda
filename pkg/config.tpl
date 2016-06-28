{
  "hostedzone": "__HOSTEDZONEID__",
  "create_ip_records": true,
  "patterns": {
    "az": "{{.Role}}.{{.AvailabilityZone}}.i.__DOMAIN__",
    "region": "{{.Role}}.{{.Region}}.i.__DOMAIN__"
  },
  "environment_name": "__ENVIRONMENT__",
  "domain": "__DOMAIN__"
}
