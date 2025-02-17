const {SQSClient, SendMessageCommand} = require("@aws-sdk/client-sqs");
const {DynamoDBClient, PutItemCommand} = require("@aws-sdk/client-dynamodb");
const { STSClient, GetCallerIdentityCommand } = require("@aws-sdk/client-sts");
const sqsClient = new SQSClient();
const dbClient = new DynamoDBClient();
const client = new STSClient();
const command = new GetCallerIdentityCommand();
const account_id = (await client.send(command)).Account;
const QUEUE_URL = `https://sqs.ap-northeast-1.amazonaws.com/${account_id}/bench-trigger-queue.fifo`;
const TABLE_NAME = "benchmark_results";

async function enqueue(id, team, endpoint, commit, actor, queued_at) {
  const body = JSON.stringify({id, team, endpoint, commit, actor, queued_at})
  const message = {
    QueueUrl: QUEUE_URL,
    MessageBody: body,
    MessageDeduplicationId: id,
    MessageGroupId: team,
  };
  const command = new SendMessageCommand(message);
  return await sqsClient.send(command);
}

async function store(id, team, endpoint, commit, actor, queued_at) {
  return await dbClient.send(new PutItemCommand({
    TableName: TABLE_NAME,
    Item: {
      id: {S: id},
      team: {S: team},
      endpoint: {S: endpoint},
      commit: commit ? {S: commit} : {NULL: true},
      bench_status: {S: "queued"},
      queued_at: {S:queued_at},
      actor: {S: actor},
    }
  }));
}

exports.handler = async function (event, context) {
  console.log("EVENT: \n" + JSON.stringify(event, null, 2));

  const qs = new URLSearchParams(event.rawQueryString);
  const queued_at = (new Date()).toISOString()
  const sqsResult = await enqueue(context.awsRequestId, qs.get("team"), qs.get("endpoint"), qs.get("commit"), qs.get("actor"), queued_at);
  console.log("SQS RESULT: \n" + JSON.stringify(sqsResult, null, 2));
  const dbResult = await store(context.awsRequestId, qs.get("team"), qs.get("endpoint"), qs.get("commit"), qs.get("actor"), queued_at);
  console.log("DynamoDB RESULT: \n" + JSON.stringify(dbResult, null, 2));

  return {
    statusCode: 200,
    body: JSON.stringify({status: "accepted"}),
  };
};
