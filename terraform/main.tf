data "template_file" "nri-ecs-task-definition-template" {
  template = file("templates/nri-ecs.json.tpl")
  vars = {
    REPOSITORY_URL = "newrelic/nri-ecs"
  }
}

resource "aws_ecs_task_definition" "nri-ecs-task-definition" {
  family                = "nri-ecs"
  execution_role_arn    = aws_iam_role.ecs-task-role.arn
  container_definitions = data.template_file.nri-ecs-task-definition-template.rendered
}

resource "aws_ecs_service" "nri-ecs-service" {
  name            = "nri-ecs"
  cluster         = aws_ecs_cluster.coreint-cluster.id
  task_definition = aws_ecs_task_definition.nri-ecs-task-definition.arn
  desired_count   = 1
}

