[![Go Report Card](https://goreportcard.com/badge/github.com/orange-car/tfplan)](https://goreportcard.com/badge/github.com/orange-car/tfplan)

# tfplan

The `tfplan` CLI is a toolbox allowing you to parse Terraform plans and perform useful actions to automate your processes. Currently, you can inspect with a filter using `tfplan inspect` to identify interesting changes. Similarly, you can compare two plans (again with a filter) to show where they diverge with `tfplan compare`.

## Stability
tfplan is currently in version 0 (unstable). Please keep this in mind and assess the risk when using the tool in critical applications. Until tfplan is released for v1, there are no stability guarantees and significant changes are expected.

## Installation
Make sure you change the v0.0.2 release number as appropriate. The below are just examples.

### Go Install
1. `go install -v github.com/orange-car/tfplan@v0.0.2` 
2. Ensure your go bin dir is in your path with `export PATH=$PATH:$HOME/go/bin`

### Script install
1. `curl https://raw.githubusercontent.com/orange-car/tfplan/master/install.sh > install.sh`
2. `./install.sh -d -r v0.0.2`

By default, this will install the tfplan binary at /usr/local/bin/tfplan. To change this, use the -b flag and specify your own directory.

Note - currently you must pass a release to the install script as tfplan is in pre-release. Once generally available, you will be able to install the "latest" release.

## Usage

### Plan Inspect
Inspects a JSON Terraform plan for changes to outputs, resource and resource drift. Compares changes against your provided filter criteria to filter out changes you don't care about. For any changes not captured by your filter, they will be printed.

Filtering Terraform plans is useful when working with Terraform programmatically such as in a deployment pipeline. For example, you may have a step to inspect the plan and explicitly exit or add a manual approval step if there are changes you are not allow-listing. A common example of this may be a change to an aws_lambda_function based on the source_code_hash. In this case, you may wish to auto-approve the Terraform plan. For any other changes, you may wish to seek manual approval before proceeding. 

Example usage:
```
$ tfplan inspect \
--plan "$(terraform show --json .plan)" \
--detailed-exitcode \
--filter "$(cat filter.json)" \
--pretty
```

In the above example, the --detailed-exitcode is used which will exit with 0 in case of no changes, 1 in case of error or 2 in case of unfiltered changes.

#### Filter Criteria
The filter criteria applies to resources, resources affected by drift and outputs. For each change type, you can specify name patterns and an array of OR conditions that can match and filter out the changing fields. Conditions for the name pattern, the path, the before and after all support * and ? wildcards. 

In this example, the criteria will filter our any change relating to resource aws_cloudwatch_log_group.this:
```
{
  "outputChanges": [],
  "driftChanges": [],
  "resourceChanges": [
    {
      "namePattern": "aws_cloudwatch_log_group.this",
      "diffs": {
        "*": [
          {
            "before": "*",
            "after": "*"
          }
        ]
      }
    }
  ]
}
```

In this example, the criteria will filter out only changes to the retention_in_days field on the resource aws_cloudwatch_log_group.this as long as the before value is 7. 
```
{
  "outputChanges": [],
  "driftChanges": [],
  "resourceChanges": [
    {
      "namePattern": "aws_cloudwatch_log_group.this",
      "diffs": {
        "retention_in_days": [
          {
            "before": "7",
            "after": "*"
          }
        ]
      }
    }
  ]
}
```

#### Sensitive, Unknown and Empty Values
The following replacements will be used for before or after values of these kinds. These replacements are matchable in your filter and not the sensitive or unknown value that it replaces.
- Empty = (empty)
- Sensitive = (sensitive value)
- unknown = (known after apply)

#### Reading the output
To filter the parsed JSON Terraform plan and work around objects of any type, tfplan will flatten object attributes (resource arguments) into a single "." separated paths with before and after values. For example, the "name" attribute for the resource aws_cloudwatch_log_group would be represented as ".name" and this is what your filter criteria needs to account for. 

For any change identified in the plan not excluded by your filter, they will be returned to you either in json format or pretty printed to the console in a style similar to Terr

### Plan Compare
Inspects two JSON Terraform plans for changes to outputs, resource and resource drift with changes filtered out by your provided filter criteria. Compares changes against each other and reports differences between the two plans.

Comparing plans programmatically is particularly useful when they are large and/or you have to do it often. Applying the optional filter can be useful to rule out changes you don't care about.

Example usage:
$ tfplan compare \
--plan-a "$(terraform show --json a.plan)" \
--plan-b "$(terraform show --json b.plan)" \
--detailed-exitcode \
--filter "$(cat filter.json)" \
--pretty

## Contributing
tfplan is open for suggestions, feedback or more direct collaboration. Feel free to open an issue or make a pull request.

### Plans
tfplan is still in early development. Read about what's planned here: [plans](docs/plans.md)

## Issues
Please open *issues* here: [New Issue](https://github.com/orange-car/tfplan/issues)