# HAWS
HAWS is an achronym from Hugo on AWS

HAWS is a simple CLI application that creates an S3 bucket and a CloudFront distribution for serving a static HUGO website.

Its current form can be used for any static site generator but only includes a configuration generator for HUGO config.

## Usage:
```  
  haws [command]

Available Commands:
  deploy      Deploy the cloudformation stacks
  generate    Generate configs
  help        Help about any command

Flags:
      --bucket-path string   Path prefix that will be appended by cloudfront to all requests (it should correspond to a sub-folder in the bucket)
      --config string        config file (default is .haws.toml in current directory)
  -h, --help                 help for haws
      --prefix string        Prefix for resources created. Can not be empty
      --record string        Record name to be added to R53 zone
      --region string        AWS region for the bucket and cloudfront distribution
      --zone-id string       AWS Id of the zone used for SSL certificate validation and where the record should be added

Use "haws [command] --help" for more information about a command.
```

### HAWS deploy
Use `haws deploy` to crate and deploy the CloudFormation templates for a new static website.

### HAWS generate
Use `haws generate` to print at the terminal the minimal config required for HUGO to use the configuration deployed earlier.
