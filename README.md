# telgram-chatops
A Telegram bot to send pipeline updates from CodePipeline to a Telegram chat

# Building
Run the following
```bash
$ make build
```

This will produce a Lambda in the `build/` directory

# Deploying
Set the following environment variables:
* `CLOUDFORMATION_BUCKET` - The Cloudformation artifact bucket
* `TELEGRAM_TOKEN` - The API token give to you by the Botfather
* `TELEGRAM_CHAT_ID` - ID of the chat to post updates to

Then run 
```bash
$ make deploy
```

This will deploy the function and an SNS topic. After this, configure CodePipeline to
send events to the topic

```bash
$ aws codestar-notifications create-notifications-rule --cli-input-json  file://rule.json
        }
```

Where `rule.json` takes the form of:
```json
{
  "Name": "CodePipeline",
  "EventTypeIds": [
    "codepipeline-pipeline-pipeline-execution-started"
  ],
  "Resource": "{Pipeline ARN}",
  "Targets": [
    {
      "TargetType": "SNS",
      "TargetAddress": "{SNS Topic ARN}"
      ],
    "Status": "ENABLED",
    "DetailType": "FULL"
}
```

`EventTypeIds` can be one of the following
* `codepipeline-pipeline-action-execution-succeeded`
* `codepipeline-pipeline-action-execution-failed`
* `codepipeline-pipeline-stage-execution-started`
* `codepipeline-pipeline-pipeline-execution-failed`
* `codepipeline-pipeline-manual-approval-failed`
* `codepipeline-pipeline-pipeline-execution-canceled`
* `codepipeline-pipeline-action-execution-canceled`
* `codepipeline-pipeline-pipeline-execution-started`
* `codepipeline-pipeline-stage-execution-succeeded`
* `codepipeline-pipeline-manual-approval-needed`
* `codepipeline-pipeline-stage-execution-resumed`
* `codepipeline-pipeline-pipeline-execution-resumed`
* `codepipeline-pipeline-stage-execution-canceled`
* `codepipeline-pipeline-action-execution-started`
* `codepipeline-pipeline-manual-approval-succeeded`
* `codepipeline-pipeline-pipeline-execution-succeeded`
* `codepipeline-pipeline-stage-execution-failed`
* `codepipeline-pipeline-pipeline-execution-superseded`

# Troubleshooting
## SNS Topic not receiving messages
In this case the pipeline's IAM role may not have permission to publish messages to the topic. Add a policy
to allow the pipeline to publish SNS messages

# Todo
* Add webhooks so the bot can answer questions about pipeline statuses
* Add webhooks to subscribe and unsubscribe via the bot itself rather than futzing with AWS