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

    // 🟢 STEP 1: เริ่มสร้าง User และส่ง Trace ID ใหม่ไปที่ Tracing Service
    const traceResponse = await axios.post("http://localhost:5001/trace", {
      service: "nodejs-app",
      operation: "create-user",
      message: `Creating user ${username}`,
    });

    const traceID = traceResponse.data.trace_id; // รับ Trace ID กลับมา
    const parentSpanID = traceResponse.data.span_id; // รับ Span ID กลับมา
    console.log(`🟢 Trace ID: ${traceID}, Parent Span ID: ${parentSpanID}`);

    // ⏳ หน่วงเวลา 500ms เพื่อจำลองกระบวนการสร้าง User
    await delay(500);

    // 🟢 STEP 2: เพิ่มข้อมูล User ลงใน Database
    console.log("🟡 Inserting data into database...");
    await axios.post("http://localhost:5001/trace", {
      trace_id: traceID, // ใช้ Trace ID เดิม
      parent_span_id: parentSpanID, // ส่ง parent_span_id ที่ได้รับจาก STEP 1
      service: "nodejs-app",
      operation: "database-insert",
      message: `Inserting user data for ${username}`,
    });

    // ⏳ หน่วงเวลา 700ms จำลองการ INSERT ข้อมูล
    await delay(5000);

    // 🟢 STEP 3: ส่งอีเมลยืนยันการสมัครสมาชิก
    console.log("🟡 Sending confirmation email...");
    await axios.post("http://localhost:5001/trace", {
      trace_id: traceID, // ใช้ Trace ID เดิม
      parent_span_id: parentSpanID, // ส่ง parent_span_id ที่ได้รับจาก STEP 2
      service: "nodejs-app",
      operation: "send-confirmation",
      message: `Sending confirmation email to ${email}`,
    });

    // ⏳ หน่วงเวลา 300ms จำลองการส่งอีเมล
    await delay(300);

    console.log("✅ User created successfully!");

    console.log("✅ User created successfully!");
    await axios.post("http://localhost:5001/trace", {
      trace_id: traceID, // ใช้ Trace ID เดิม
      parent_span_id: parentSpanID, // ส่ง parent_span_id ที่ได้รับจาก STEP 2
      service: "nodejs-app",
      operation: "show user created successfully",
      message: ` User created successfully! ${email} ${username}`,
    });
    res.json({ message: "User created successfully!", trace_id: traceID });
  } catch (error) {
    console.error("❌ Error sending trace:", error);
    res.status(500).json({ error: "Failed to send trace" });
  }
});

app.listen(PORT, () => {
  console.log(`🚀 Node.js app running at http://localhost:${PORT}`);
});
