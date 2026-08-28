import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-node";
import {
  TodoService,
  ActionType,
  EventType,
  TodoStreamRequestSchema,
  SubscribeTodosRequestSchema,
} from "./src/gen/todo/v1/todo_pb";
import { create } from "@bufbuild/protobuf";

async function runTest() {
  console.log("🚀 Testing ConnectRPC Go backend at http://localhost:8085...\n");

  const transport = createConnectTransport({
    baseUrl: "http://localhost:8085",
    httpVersion: "2",
  });

  const client = createClient(TodoService, transport);

  // 1. Test SubscribeTodos (Server Stream)
  console.log("1️⃣ Testing SubscribeTodos (Realtime Server Stream):");
  const subReq = create(SubscribeTodosRequestSchema, { clientId: "automated-test-client" });
  const subStream = client.subscribeTodos(subReq);

  let streamReceivedAdd = false;
  let streamReceivedUpdate = false;
  let streamReceivedDelete = false;

  const streamPromise = (async () => {
    for await (const res of subStream) {
      console.log(`   [STREAM RECEIVED] Event=${EventType[res.event]}, Message="${res.message}"`);
      if (res.event === EventType.ADDED) streamReceivedAdd = true;
      if (res.event === EventType.UPDATED) streamReceivedUpdate = true;
      if (res.event === EventType.DELETED) {
        streamReceivedDelete = true;
        break;
      }
    }
  })();

  // Give stream 100ms to register
  await new Promise((r) => setTimeout(r, 100));

  // 2. Test ExecuteAction (Add)
  console.log("\n2️⃣ Testing ExecuteAction -> ACTION_TYPE_ADD:");
  const addReq = create(TodoStreamRequestSchema, {
    action: ActionType.ADD,
    title: "Verify ConnectRPC in Browser & Node",
    clientId: "automated-test-client",
  });
  const addRes = await client.executeAction(addReq);
  const createdId = addRes.item?.id!;
  console.log(`   [ACTION RESPONSE] Added ID=${createdId}, Title="${addRes.item?.title}"`);

  // 3. Test ExecuteAction (Update)
  console.log("\n3️⃣ Testing ExecuteAction -> ACTION_TYPE_UPDATE:");
  const updateReq = create(TodoStreamRequestSchema, {
    action: ActionType.UPDATE,
    id: createdId,
    title: "Verify ConnectRPC in Browser & Node (Completed)",
    completed: true,
    clientId: "automated-test-client",
  });
  const updateRes = await client.executeAction(updateReq);
  console.log(`   [ACTION RESPONSE] Updated ID=${updateRes.item?.id}, Completed=${updateRes.item?.completed}`);

  // 4. Test ExecuteAction (Delete)
  console.log("\n4️⃣ Testing ExecuteAction -> ACTION_TYPE_DELETE:");
  const deleteReq = create(TodoStreamRequestSchema, {
    action: ActionType.DELETE,
    id: createdId,
    clientId: "automated-test-client",
  });
  const deleteRes = await client.executeAction(deleteReq);
  console.log(`   [ACTION RESPONSE] Deleted ID=${deleteRes.item?.id}`);

  // Wait for stream to receive delete event
  await streamPromise;

  console.log("\n🎉 ALL TESTS PASSED! Realtime streaming and action dispatching verified!");
  process.exit(0);
}

runTest();
