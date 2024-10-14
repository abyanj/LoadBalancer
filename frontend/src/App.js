import React, { useEffect, useState } from 'react';
import { FaServer } from 'react-icons/fa';

function App() {
  const [serverStatus, setServerStatus] = useState([]);
  const [requestLog, setRequestLog] = useState([]); // updated to log

  useEffect(() => {
    let ws;

    const connectWebSocket = () => {
      ws = new WebSocket('ws://localhost:8080/ws');

      ws.onopen = () => {
        console.log("Connected to WebSocket server");
      };

      ws.onmessage = (event) => {
        console.log("Received WebSocket message:", event.data);
        const update = JSON.parse(event.data);

        if (update.MessageType === "status") {
          
          setServerStatus((prevStatus) => {
            const index = prevStatus.findIndex(s => s.Server === update.Server);
            if (index === -1) {
              return [...prevStatus, update];
            } else {
              const updatedStatus = [...prevStatus];
              updatedStatus[index] = update;
              return updatedStatus;
            }
          });
        } else if (update.MessageType === "request") {
          // no duplicate
          setRequestLog((prevLog) => {
            if (prevLog.find((log) => log.Timestamp === update.Timestamp)) {
              return prevLog; 
            }
            return [update, ...prevLog]; 
          });
        }
      };

      ws.onerror = (error) => {
        console.error("WebSocket error:", error);
      };

      ws.onclose = (event) => {
        console.log(`WebSocket closed: ${event.code} (${event.reason})`);
        if (event.code === 1006) {
          console.log("Abnormal closure detected, retrying...");
          setTimeout(connectWebSocket, 3000); // Try to reconnect after 3 seconds
        }
      };
    };

    connectWebSocket();

    return () => {
      ws.close();
    };
  }, []);

  return (
    <div className="min-h-screen bg-black text-gray-100 p-6">
      <header className="text-center mb-8">
        <h1 className="text-5xl font-bold text-gray-100">Load Balancer Dashboard</h1>
        <p className="text-base text-gray-400">Monitor the health and routing of your backend servers</p>
      </header>

      {/* Server Status Section */}
      <section className="mb-10">
        <h2 className="text-3xl font-semibold text-gray-300 mb-6">Server Status</h2>
        <div className="grid grid-cols-3 gap-8">
          {serverStatus.map((status, index) => (
            <div
              key={index}
              className={`flex items-center p-8 bg-gray-800 shadow-lg rounded-lg ${status.Healthy ? 'border-green-400' : 'border-red-400'} border`}
            >
              <FaServer className="text-4xl mr-5 text-gray-400" />
              <div>
                <h3 className="text-2xl font-semibold">{status.Server}</h3>
                <div className="flex items-center mt-2">
                  <span className={`inline-block w-4 h-4 rounded-full mr-2 ${status.Healthy ? 'bg-green-500' : 'bg-red-500'}`}></span>
                  <p className={status.Healthy ? 'text-green-400' : 'text-red-400'}>
                    {status.Healthy ? 'Online' : 'Offline'}
                  </p>
                </div>
              </div>
            </div>
          ))}
        </div>
      </section>

      {/* Current Request Log Section */}
      <section className="mb-10">
        <h2 className="text-3xl font-semibold text-gray-300 mb-6">Request Log</h2>
        <div className="space-y-4">
          {requestLog.length === 0 ? (
            <p className="text-gray-500 text-lg">No requests logged yet.</p>
          ) : (
            requestLog.map((request, index) => (
              <div
                key={index}
                className="p-5 bg-gray-800 shadow-md border border-blue-500 rounded-lg"
              >
                <h3 className="text-lg font-semibold text-blue-400">{`Request forwarded to: ${request.Server}`}</h3>
                {request.Skipped && (
                  <p className="text-md text-yellow-400">{`Skipped unhealthy server: ${request.Skipped}`}</p>
                )}
                <p className="text-sm text-gray-400">{`Timestamp: ${request.Timestamp}`}</p>
              </div>
            ))
          )}
        </div>
      </section>
    </div>
  );
}

export default App;
