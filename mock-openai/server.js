const http = require("http");

const server = http.createServer((req, res) => {
  if (req.method === "POST" && req.url === "/chat/completions") {
    let body = "";
    req.on("data", (chunk) => (body += chunk));
    req.on("end", () => {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(
        JSON.stringify({
          id: "chatcmpl-mock",
          object: "chat.completion",
          created: Math.floor(Date.now() / 1000),
          model: "gpt-4o-mini",
          choices: [
            {
              index: 0,
              message: {
                role: "assistant",
                content:
                  "This is a mock response from the test OpenAI server.",
              },
              finish_reason: "stop",
            },
          ],
          usage: { prompt_tokens: 10, completion_tokens: 10, total_tokens: 20 },
        })
      );
    });
  } else {
    res.writeHead(404, { "Content-Type": "application/json" });
    res.end(JSON.stringify({ error: "Not found" }));
  }
});

const port = process.env.PORT || 4010;
server.listen(port, "0.0.0.0", () => {
  console.log(`Mock OpenAI server listening on port ${port}`);
});
