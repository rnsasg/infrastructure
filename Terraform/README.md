# Terraform (Infrastructure as Code)

## Code Structure best practices

- `main.tf` - call modules, locals, and data sources to create all resources
- `variables.tf` - contains declarations of variables used in `main.tf`
- `outputs.tf` - contains outputs from the resources created in `main.tf`
- `versions.tf` - contains version requirements for Terraform and pro