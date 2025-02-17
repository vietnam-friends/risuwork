async function notify(message) {
  const response = await fetch("https://slack.com/api/chat.postMessage", {
    method: "POST",
    headers: {
      Authorization: `Bearer ${process.env.SLACK_TOKEN}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      channel: `recruit-isucon-2024-${message.team}`,
      blocks: [
        {
          type: "header",
          text: {
            type: "plain_text",
            text: message.score > 0 ? `:rocket: ベンチ結果: ${message.score.toLocaleString()}` : ":sob: ベンチ結果: 0"
          }
        },
        {
          type: "section",
          fields: [
            {
              type: "mrkdwn",
              text: message.lang ? `lang: *${message.lang}*` : "lang: -"
            },
            {
              type: "mrkdwn",
              text: message.team && message.commit ? `commit: *<https://github.com/risuwork/${message.team}/commit/${message.commit}|${message.commit.slice(0, 7)}>*` : "commit: -"
            },
            {
              type: "mrkdwn",
              text: message.actor ? `actor: *${message.actor}*` : "actor: -"
            },
            {
              type: "mrkdwn",
              text: message.queued_at ? `queued: *<!date^${Math.round(Date.parse(message.queued_at) / 1000)}^{time_secs}|${message.queued_at}>*` : "queued: -"
            },
            {
              type: "mrkdwn",
              text: message.started_at ? `start: *<!date^${Math.round(Date.parse(message.started_at) / 1000)}^{time_secs}|${message.started_at}>*` : "start: -"
            },
            {
              type: "mrkdwn",
              text: message.ended_at ? `end: *<!date^${Math.round(Date.parse(message.ended_at) / 1000)}^{time_secs}|${message.ended_at}>*` : "end: -"
            }
          ]
        },
        {
          type: "context",
          elements: [
            {
              type: "mrkdwn",
              text: `bench ID: ${message.id}`
            }
          ]
        }
      ],
      username: "bench-notifier",
      icon_emoji: ":squirrel:",
    }),
  });
  const result = await response.json();
  if (!result.ok) {
    console.error(result);
  }

  const threadResponse = await fetch("https://slack.com/api/chat.postMessage", {
    method: "POST",
    headers: {
      Authorization: `Bearer ${process.env.SLACK_TOKEN}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      channel: `recruit-isucon-2024-${message.team}`,
      thread_ts: result.ts,
      blocks: [
        {
          type: "rich_text",
          elements: [
            {
              type: "rich_text_section",
              elements: [
                {
                  type: "text",
                  text: "メッセージ\n",
                  style: {
                    italic: true
                  }
                }
              ]
            },
            {
              type: "rich_text_list",
              style: "bullet",
              elements: (message.messages.map(m => ({
                type: "rich_text_section",
                elements: [{"type": "text", "text": m}]
              })))
            }
          ]
        },
      ],
      username: "bench-notifier",
      icon_emoji: ":squirrel:",
    }),
  });
  const threadResult = await threadResponse.json();
  if (!threadResult.ok) {
    console.error(threadResult);
  }

}

exports.handler = async function (event, context) {
  console.debug("EVENT: \n" + JSON.stringify(event, null, 2));

  const messages = event.Records.map(r => JSON.parse(r.Sns.Message));

  await Promise.all(messages.map(notify));
};
