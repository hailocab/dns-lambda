{
  "hostedzone": "__HOSTEDZONEID__",
  "patterns": {
    "az": "{{.AutoScalingGroup}}.{{.AvailabilityZone}}.i.{{.EnvironmentName}}.__DOMAIN__",
    "region": "{{.AutoScalingGroup}}.{{.Region}}.i.{{.EnvironmentName}}.__DOMAIN__"
  },
  "environment_name": "__ENVIRONMENT__"
}
