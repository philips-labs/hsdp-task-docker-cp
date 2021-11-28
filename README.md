# hsdp-task-docker-cp

A task for use with [hsdp_function](https://registry.terraform.io/providers/philips-software/hsdp/latest/docs/resources/function) 
to synchronize HSDP Docker namespaces across regions

# Usage

```hcl
module "siderite_backend" {
  source = "philips-labs/siderite-backend/cloudfoundry"

  cf_region   = "eu-west"
  cf_org_name = "hsdp-demo-org"
  cf_user     = var.cf_user
  iron_plan = "medium-encrypted"
}

resource "hsdp_function" "docker_cp" {
  name         = "hsdp-docker-cp"
  docker_image = "philipslabs/hsdp-task-docker-cp:latest"
  command      = ["hsdp-docker-cp"]

  environment = {
    # Source
    CP_SOURCE_HOST      = "docker.na1.hsdp.io"
    CP_SOURCE_LOGIN     = "cf-functional-account-na1"
    CF_SOURCE_PASSWORD  = "passw0rdH3r3"
    CF_SOURCE_NAMESPACE = "loafoe"  
    
    # Destination
    CF_DEST_HOST        = "docker.eu1.hsdp.io"
    CF_DEST_LOGIN       = "cf-functional-account-eu1"
    CF_DEST_PASSWORD    = "An0therpAssw0rd"
  }

  # Run every 60m
  run_every = "60m"

  # Run for max 60 minutes at a time
  timeout = 3600

  backend {
    credentials = module.siderite_backend.credentials
  }
}
```

# Contact / Getting help

Please post your questions on the HSDP Slack `#terraform` channel

# License

License is MIT
