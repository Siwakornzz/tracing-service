const express = require("express");
const axios = require("axios");

const app = express();
const PORT = 3000;

app.use(express.json());

const delay = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

app.get("/test-trace", async (req, res) => {
  const username = "aun";
  const email = "aun@gmail.com";

  try {
    console.log("🚀 Start tracing process...");

    let startTime = new Date().toISOString();
    await delay(500);
    let endTime = new Date().toISOString();

    // Start Root Trace
    const traceResponse = await axios.post(
      "http://localhost:5001/start-trace",
      {
        service: "nodejs-app",
        operation: "create-user",
        message: `Creating user ${username}`,
        start_time: startTime,
      }
    );

    const traceID = traceResponse.data.trace_id;
    const parentSpanID = traceResponse.data.span_id;
    console.log(`🟢 Trace ID: ${traceID}`);

    // Nested: Insert into Database
    startTime = new Date().toISOString();
    await delay(500);
    endTime = new Date().toISOString();
    const dbResponse = await axios.post("http://localhost:5001/add-trace", {
      trace_id: traceID,
      parent_span_id: parentSpanID,
      service: "nodejs-app",
      operation: "database-insert",
      message: `Inserting user data for ${username}`,
      start_time: startTime,
    });
    const dbSpanID = dbResponse.data.span_id;

    // Nested: Send Confirmation Email
    startTime = new Date().toISOString();
    await delay(700);
    endTime = new Date().toISOString();
    const emailResponse = await axios.post("http://localhost:5001/add-trace", {
      trace_id: traceID,
      parent_span_id: dbSpanID,
      service: "nodejs-app",
      operation: "send-confirmation",
      message: `Sending confirmation email to ${email}`,
      start_time: startTime,
    });
    const emailSpanID = emailResponse.data.span_id;

    // Stop spans
    await axios.post("http://localhost:5001/stop-trace", {
      span_id: emailSpanID,
      end_time: new Date().toISOString(),
    });
    await axios.post("http://localhost:5001/stop-trace", {
      span_id: dbSpanID,
      end_time: new Date().toISOString(),
    });
    await axios.post("http://localhost:5001/stop-trace", {
      span_id: parentSpanID,
      end_time: new Date().toISOString(),
    });

    console.log("✅ User created successfully!");
    res.json({ message: "User created successfully!", trace_id: traceID });
  } catch (error) {
    console.error("❌ Error sending trace:", error);
    res.status(500).json({ error: "Failed to send trace" });
  }
});

app.listen(PORT, () => {
  console.log(`🚀 Node.js app running at http://localhost:${PORT}`);
});
