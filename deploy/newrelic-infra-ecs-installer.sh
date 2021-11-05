#!/usr/bin/env bash

export AWS_PAGER=""
DRY_RUN=""

set -o errexit
set -o pipefail
set -o nounset
#set -o xtrace

usage() {
  cat <<-HELPMSG
New Relic ECS integration installer.

The following tools are required:

  - awscli https://aws.amazon.com/cli/
  - grep
  - awk
  - tr
  - curl or wget

The installer executes the following steps:

  - Creates the required AWS resources. (All launch types)

  - Downloads the task definition template hosted in
    https://download.newrelic.com/infrastructure_agent/integrations/ecs/newrelic-infra-ecs-latest.json
    saving it as "newrelic-infra-task-definition.json". (EC2/EXTERNAL launch type only)

  - Replaces the placeholders in the template. (EC2/EXTERNAL launch type only)

  - Registers the task definition from "newrelic-infra-task-definition.json". (EC2/EXTERNAL launch type only)

  - Creates a service for the registered task for EC2 launch type unless -e is defined 
    where EXTERNAL launch type service is created.

The installation process creates the task definition file
"newrelic-infra-task-definition.json". When this file is present the
installer won't try to create/update any ECS resource other than the ECS task
and service. If you want the installer to try creating all the other AWS
resources just delete "newrelic-infra-task-definition.json" and
execute the command again.

When creating AWS resources the region configured in your awscli profile will
be used, to check the region you have set up run the following command:

  $ aws configure get region
  us-east-1

Executing the installer with the default values creates the following AWS
resources:

  - Systems Manager (SSM) parameter "/newrelic-infra/ecs/license-key". (All launch types)

  - IAM policy "NewRelicSSMLicenseKeyReadAccess" which enables access to the
    SSM parameter with the license key. (All launch types)

  - IAM role "NewRelicECSTaskExecutionRole" to be used as the task execution role
    https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_execution_IAM_role.html
    Policies attached to the role (All launch types):

      * NewRelicSSMLicenseKeyReadAccess (created by the installer).
      * AmazonEC2ContainerServiceforEC2Role
      * AmazonECSTaskExecutionRolePolicy

  - Registers the "newrelic-infra" ECS task definition. (EC2/EXTERNAL launch type)

  - Creates the service "newrelic-infra" for the registered task
    using DAEMON scheduling strategy. (EC2 launch type)

USAGE:
  $0 ARGS [OPTIONS]

ARGS:
  -c Cluster name.
  -l New Relic license key. Not required for uninstall.
  -n Task definition family name (defaults to newrelic-infra).

OPTIONS:
  -h  Print help information.
  -u  Uninstall. Deletes all the AWS resources related to the integration. The
      same arguments used for installing the integration must be specified
      when running the uninstall command, otherwise the default names will be
      used when looking what resources to delete.
  -d  Dry run (only valid when uninstall -u is set). Print all commands that will be executed without running them.
      The aws cli is used for validation, so an aws token is required.
  -f  Fargate mode (defaults to false)
  -e  Create External ECS instance Daemon Service.
HELPMSG
}

uninstall() {
  local license_key_parameter=$1
  local role=$2
  local policy=$3
  local service=$4
  local service_external=$5
  local cluster=$6
  local task_family_name=$7

  if [ $DRY_RUN ]; then
    echo "Dry run set, printing aws commands"
  fi

  echo "Deleting parameter $license_key_parameter..."
  local license_key_parameter_arn
  if license_key_parameter_arn=$(get_license_key_parameter_arn "$license_key_parameter"); then
    run aws ssm delete-parameter --name "$license_key_parameter"
  else
    echo "Parameter $license_key_parameter not found, skipping"
  fi

  echo "Deleting role $role..."
  local policy_arns
  policy_arns=$(list_attached_role_policies "$role")
  for arn in $(split_spaces "$policy_arns"); do
    run aws iam detach-role-policy --role-name "$role" --policy-arn "$arn"
  done

  run aws iam delete-role --role-name "$role"

  echo "Deleting policy $policy..."
  local policy_arn
  local policy_versions
  if policy_arn=$(get_policy_arn "$policy") && [ -n "$policy_arn" ]; then
    policy_versions=$(list_non_default_policy_versions "$policy_arn")
    for version in $(split_spaces "$policy_versions"); do
      run aws iam delete-policy-version --policy-arn "$policy_arn" --version-id "$version"
    done
    run aws iam delete-policy --policy-arn "$policy_arn"
  else
    echo "Policy $policy not found, skipping"
  fi

  echo "Deleting service $service..."
  run aws ecs delete-service --service "$service" --cluster "$cluster"
  
  echo "Deleting service $service..."
  run aws ecs delete-service --service "$service_external" --cluster "$cluster"

  echo "Deregistering all newrelic-infra task definitions..."
  local task_arns
  task_arns=$(aws ecs list-task-definitions \
    --family-prefix "$task_family_name" \
    --output text \
    --query taskDefinitionArns)
  for arn in $(split_spaces "$task_arns"); do
    run aws ecs deregister-task-definition --task-definition "$arn"
  done
}

split_spaces() {
  echo "$1" | tr " " "\n"
}

list_non_default_policy_versions() {
  local policy_arn=$1
  aws iam list-policy-versions \
    --policy-arn "${policy_arn}" \
    --output text \
    --query 'Versions[*].[VersionId,IsDefaultVersion]' |
    grep False |
    awk '{print $1}' 2>/dev/null
}

list_attached_role_policies() {
  local role=$1
  aws iam list-attached-role-policies \
    --role-name "${role}" \
    --output text \
    --query 'AttachedPolicies[*].PolicyArn' 2>/dev/null
}

get_execution_role_arn() {
  local role_name=$1
  aws iam get-role --role-name "$role_name" --output text --query Role.Arn 2>/dev/null
  return $?
}

valid_args() {
  for var in "$@"; do
    if [ -z "$var" ]; then
      return 1
    fi
  done
}

get_license_key_parameter_arn() {
  local parameter_name=$1
  aws ssm get-parameter \
    --name "$parameter_name" \
    --output text \
    --query Parameter.ARN \
    2>/dev/null
}

create_license_key_parameter() {
  local license_key=$1
  local parameter_name=$2
  aws ssm put-parameter \
    --name "$parameter_name" \
    --type SecureString \
    --description 'New Relic license key for ECS monitoring' \
    --value "$license_key" >>/dev/null
}

function error_exit() {
  echo "Error: ${1:-"Unknown Error"}" 1>&2
  exit 1
}

retrieve_url() {
  local url=$1
  local destination=$2

  if command -v curl >>/dev/null 2>&1; then
    curl --fail -s -o "$destination" "$url"
  elif command -v wget >>/dev/null 2>&1; then
    wget -q -O "$destination" "$url"
  else
    error_exit "Unable to find either curl or wget"
  fi
}

check_required_tools() {

  if ! command -v aws >>/dev/null 2>&1; then
    echo "awscli not found in the system. Get it from https://aws.amazon.com/cli/"
    exit 1
  fi

  system_tools=(awk grep tr)
  for tool in "${system_tools[@]}"; do
    if ! command -v "$tool" >>/dev/null 2>&1; then
      echo "$tool not found in the system"
      exit 1
    fi
  done

  command -v curl >>/dev/null 2>&1
  local curl_found=$?
  command -v wget >>/dev/null 2>&1
  local wget_found=$?
  if [ $curl_found -ne 0 ] && [ $wget_found -ne 0 ]; then
    echo "Neither curl nor wget were found in the system, install one of them"
    exit 1
  fi
}

register_task_definition() {
  local task_definition_file=$1
  aws ecs register-task-definition --cli-input-json file://"${task_definition_file}" >>/dev/null
}

create_service() {
  local cluster_name=$1
  local service_name=$2
  local task_definition=$3
  local launch_type=$4
  aws ecs create-service \
    --cluster "$cluster_name" \
    --service-name "$service_name" \
    --task-definition "$task_definition" \
    --scheduling-strategy DAEMON \
    --launch-type "$launch_type"  >>/dev/null
}

set_task_family_name() {
  local task_definition_file=$1
  local task_family_name=$2
  sed -i.bak "s|\"family\": \"newrelic-infra\"|\"family\": \"${task_family_name}\"|" "$task_definition_file"
  local return=$?
  rm "${task_definition_file}.bak" >>/dev/null 2>&1 || true
  return $return
}

set_task_execution_role() {
  local task_definition_file=$1
  local task_execution_role_arn=$2
  sed -i.bak "s|<YOUR_TASK_EXECUTION_ROLE>|${task_execution_role_arn}|" "$task_definition_file"
  local return=$?
  rm "${task_definition_file}.bak" >>/dev/null 2>&1 || true
  return $return
}

set_task_license_key_parameter_name() {
  local task_definition_file=$1
  local license_key_parameter_name=$2
  sed -i.bak "s|<SYSTEM_MANAGER_LICENSE_PARAMETER_NAME>|${license_key_parameter_name}|" "$task_definition_file"
  local return=$?
  rm "${task_definition_file}.bak" >>/dev/null 2>&1 || true
  return $return
}

set_task_deploy_method() {
  local task_definition_file=$1
  sed -i.bak "s|downloadPage|installScript|" "$task_definition_file"
  local return=$?
  rm "${task_definition_file}.bak" >>/dev/null 2>&1 || true
  return $return
}

service_exists() {
  local cluster_name=$1
  local service_name=$2
  aws ecs list-services --cluster "$cluster_name" --output text | grep "$service_name$" >>/dev/null
}

update_service() {
  local cluster_name=$1
  local service_name=$2
  local task_definition=$3
  local launch_type=$4
  aws ecs update-service --cluster "$cluster_name" --service "$service_name" --task-definition "$task_definition" --launch-type "$launch_type" >>/dev/null
}

create_task_execution_role() {
  local role_name=$1

  aws iam create-role \
    --role-name "$role_name" \
    --assume-role-policy-document '{"Version":"2008-10-17","Statement":[{"Sid":"","Effect":"Allow","Principal":{"Service":"ecs-tasks.amazonaws.com"},"Action":"sts:AssumeRole"}]}' \
    --description "ECS task execution role for New Relic infrastructure" \
    --output text \
    --query Role.Arn
  return $?
}

create_policy_for_license_key_access() {
  local policy_name=$1
  local license_key_arn=$2

  aws iam create-policy \
    --policy-name "$policy_name" \
    --policy-document "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Action\":[\"ssm:GetParameters\"],\"Resource\":[\"${license_key_arn}\"]}]}" \
    --description "Provides read access to the New Relic SSM license key parameter" \
    --output text \
    --query Policy.Arn
  return $?
}

get_policy_arn() {
  local policy_arn=$1
  aws iam list-policies --query "Policies[?PolicyName==\`$policy_arn\`].Arn" --output text
  return $?
}

attach_policies_to_role() {
  local role=$1
  shift
  echo "$@" |
    xargs -n1 aws iam attach-role-policy \
      --role-name "$role" \
      --policy-arn || return 1
}

run() {
  if [ "$DRY_RUN" ]; then
    echo "> $*"
    return 0
  fi

  eval "$@"
}

create_or_get_license_key_parameter_arn() {
  local license_key_parameter_name=$1
  local license_key_parameter_arn
  if ! license_key_parameter_arn="$(get_license_key_parameter_arn "$license_key_parameter_name")"; then
    echo "License key parameter not found. Creating..." >&2
    create_license_key_parameter "$license_key" "$license_key_parameter_name" || error_exit "couldn't create license key parameter"
    license_key_parameter_arn="$(get_license_key_parameter_arn "$license_key_parameter_name")" || error_exit "couldn't retrieve ARN for license key parameter"
  else
    echo "License key parameter found." >&2
  fi
  echo "$license_key_parameter_arn"
}

create_or_get_task_execution_role() {
  local task_execution_role=$1
  local license_key_access_policy_name=$2
  local license_key_parameter_arn=$3

  local task_execution_role_arn
  if ! task_execution_role_arn=$(get_execution_role_arn "$task_execution_role"); then
    echo "Task execution role $task_execution_role not found. Creating..." >&2
    task_execution_role_arn=$(create_task_execution_role "$task_execution_role") || error_exit "couldn't create role $task_execution_role"

    local policy_arn
    policy_arn=$(get_policy_arn "$license_key_access_policy_name") || error_exit "failed to get policy arn"
    if [ -z "$policy_arn" ]; then
      echo "Creating policy for New Relic license key access..." >&2
      policy_arn=$(create_policy_for_license_key_access "$license_key_access_policy_name" "$license_key_parameter_arn") || error_exit "couldn't create policy for access to the license key parameter"
    fi

    policies_arns+=("$policy_arn")
    echo "Attaching the following policies to the role $task_execution_role..." >&2
    printf '  %s\n' "${policies_arns[@]}" >&2
    attach_policies_to_role "$task_execution_role" "${policies_arns[@]}" || error_exit "couldn't attach policies to the role $task_execution_role"
  fi

  echo "$task_execution_role_arn"
}

main() {
  local cluster_name=""
  local license_key=""
  local fargate=false
  local external=false
  local is_uninstall=false
  local task_execution_role="NewRelicECSTaskExecutionRole"
  local license_key_parameter_name="/newrelic-infra/ecs/license-key"
  local license_key_access_policy_name="NewRelicSSMLicenseKeyReadAccess"
  local policies_arns=("arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role" "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy")
  local task_definition_url="https://download.newrelic.com/infrastructure_agent/integrations/ecs/newrelic-infra-ecs-ec2-latest.json"
  local task_definition_file="newrelic-infra-task-definition.json"
  local service_name="newrelic-infra"
  local service_name_launch_type="EC2"
  local service_name_external="newrelic-infra-external"
  local service_name_external_launch_type="EXTERNAL"
  local task_family_name="newrelic-infra"

  while getopts c:l:n:hudfe option; do
    case "$option" in
    l)
      license_key=$OPTARG
      ;;
    c)
      cluster_name=$OPTARG
      ;;
    h)
      usage
      exit 0
      ;;
    u)
      is_uninstall=true
      ;;
    d)
      DRY_RUN=true
      ;;
    e)
      external=true
      ;;
    f)
      fargate=true
      ;;
    n)
      task_family_name=$OPTARG
      ;;
    *) ;;
    esac
  done

  if [ "$is_uninstall" == true ]; then
    echo "Uninstalling..."
    if ! valid_args "$cluster_name"; then
      usage
      exit 1
    fi

    uninstall \
      "$license_key_parameter_name" \
      "$task_execution_role" \
      "$license_key_access_policy_name" \
      "$service_name" \
      "$service_name_external" \
      "$cluster_name" \
      "$task_family_name"
    exit 0
  fi

  if ! valid_args "$license_key" "$cluster_name"; then
    usage
    exit 1
  fi

  echo "Installing the New Relic ECS integration..."
  if [ "$fargate" == true ]; then
    echo "Detected Fargate launch type parameter."
  else
    if [ "$external" == true ]; then
      echo "Detected EXTERNAL launch type parameter."
    else
      echo "Detected EC2 launch type (default)."
    fi
  fi

  echo "Checking required tools."
  check_required_tools

  if [ "$fargate" == true ]; then
    echo "Checking if license key parameter exists in the AWS parameter store."

    local license_key_parameter_arn
    license_key_parameter_arn="$(create_or_get_license_key_parameter_arn "$license_key_parameter_name")" || exit 1

    task_execution_role_arn=$(create_or_get_task_execution_role "$task_execution_role" \
      "$license_key_access_policy_name" \
      "$license_key_parameter_arn") || exit 1

    echo "Finished installation of the ECS integration's AWS resources successfully."
    return
  fi

  if [ -f $task_definition_file ]; then
    echo "Found local task definition file $task_definition_file."
    echo "  Task execution role and license key parameter won't be replaced in the existing task definition file."
    echo "  This tool will only update the ecs service while an existing task definition file is present."

    read -p "Do you want to remove it? [y/n] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
      echo "Removing existing $task_definition_file."
      rm $task_definition_file
    fi
  fi

  if [ ! -f $task_definition_file ]; then
    echo "Downloading task definition file task_definition_file from:"
    echo "  $task_definition_url"

    retrieve_url "$task_definition_url" $task_definition_file || error_exit "couldn't download task definition."
    set_task_deploy_method $task_definition_file || error_exit "couldn't set the deploy method on the task definition."
    echo "Checking if license key parameter exists in the AWS parameter store."

    local license_key_parameter_arn
    license_key_parameter_arn="$(create_or_get_license_key_parameter_arn "$license_key_parameter_name")" || exit 1

    task_execution_role_arn=$(create_or_get_task_execution_role "$task_execution_role" \
      "$license_key_access_policy_name" \
      "$license_key_parameter_arn") || exit 1

    set_task_execution_role "$task_definition_file" "$task_execution_role_arn"
    set_task_license_key_parameter_name "$task_definition_file" \
      "$license_key_parameter_name" || error_exit "couldn't set the license key parameter in the task definition."
  fi

  set_task_family_name "$task_definition_file" "$task_family_name"
  local task_definition
  task_definition=$(awk 'BEGIN { FS="\""; RS="," }; { if ($2 == "family") {print $4} }' "$task_definition_file")
  if [ -z "$task_definition" ]; then
    error_exit "Task family not found on definition file $task_definition_file"
  fi

  echo "Registering task definition $task_definition from file $task_definition_file..."
  register_task_definition "$task_definition_file" || error_exit "couldn't register task definition $task_definition_file."

  if [ "$external" == true ]; then
    echo "Checking if service $service_name_external in cluster $cluster_name already exists..."

    if service_exists "$cluster_name" "$service_name_external"; then
      echo "Service $service_name_external in cluster $cluster_name already exists. Updating to latest task definition of $task_definition."
      update_service "$cluster_name" "$service_name_external" "$task_definition" "$service_name_external_launch_type" || error_exit "couldn't update the $service_name_external service."
    else
      echo "Creating daemon service $service_name_external for task $task_definition in cluster $cluster_name..."
      create_service "$cluster_name" "$service_name_external" "$task_definition" "$service_name_external_launch_type" || error_exit "couldn't create the $service_name_external service."
    fi
  else # EC2 default launch type.
    echo "Checking if service $service_name in cluster $cluster_name already exists..."

    if service_exists "$cluster_name" "$service_name"; then
      echo "Service $service_name in cluster $cluster_name already exists. Updating to latest task definition of $task_definition."
      update_service "$cluster_name" "$service_name" "$task_definition" "$service_name_launch_type" || error_exit "couldn't update the $service_name service."
    else
      echo "Creating daemon service $service_name for task $task_definition in cluster $cluster_name..."
      create_service "$cluster_name" "$service_name" "$task_definition" "$service_name_launch_type" || error_exit "couldn't create the $service_name service."
    fi
  fi

  echo "Finished installation of the ECS integration successfully."
}

main "$@" || exit 1
