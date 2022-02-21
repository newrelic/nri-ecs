[
  {
    "essential": true,
    "memory": 256,
    "name": "nri-ecs",
    "cpu": 256,
    "image": "${REPOSITORY_URL}:latest",
    "environment": [
      {"name": "ENABLE_NRI_ECS", "value": "true"},
      {"name": "NRIA_PASSTHROUGH_ENVIRONMENT", "value": "ECS_CONTAINER_METADATA_URI,ENABLE_NRI_ECS"},
      {"name": "NRIA_VERBOSE", "value": "1"},
      {"name": "NRIA_LICENSE_KEY", "value": "arn:aws:secretsmanager:eu-central-1:801306408012:secret:NewRelicLicenseKeySecret-UTOm7ZmV4V8p-MgdGow"}
    ]
  }
]
