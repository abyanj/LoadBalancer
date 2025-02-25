# Golang Load Balancer

A simple **application-layer (Layer 7) HTTP load balancer** built using **Golang**. It distributes incoming traffic across multiple backend servers, ensuring **efficient load distribution, high availability, and automatic health checks**. The project also includes a **real-time monitoring dashboard** built with **React.js & WebSockets**, providing live updates on request routing and server health.

![Alt text](/loadbalancer.png "Load Balancer")
---

## 🚀 Features
- **Round-robin load balancing**: Evenly distributes requests among backend servers.
- **Health checks**: Periodically checks backend servers and removes unhealthy ones.
- **WebSocket-based real-time dashboard**: Displays:
  - Live server health statuses (🟢 Online / 🔴 Offline).
  - Requests being routed and skipped servers.
- **Automatic server recovery detection**: Once a failed server comes back online, traffic is routed to it again.
- **Play button to start servers**: Click on a server to attempt restarting it via the frontend.

---

## 🛠️ Technologies Used
### **Backend (Load Balancer)**
- **Golang** (HTTP handling, concurrency, WebSockets)
- **Gorilla WebSocket** (Real-time communication)
- **Sync/atomic & Mutexes** (Safe concurrent operations)

### **Frontend (Dashboard)**
- **React.js** (UI framework)
- **WebSockets** (Live server updates)
- **TailwindCSS** (Dark-themed, responsive UI)
- **React Icons** (Server/play button icons)

---

## 📌 Architecture Overview
1. **Client makes a request** → **Load balancer (`localhost:8080/`) receives it**.
2. **Load balancer selects the next healthy backend server** and forwards the request.
3. **Backend server processes the request** and returns a response.
4. **Load balancer forwards the response** back to the client.
5. **Frontend (React.js) receives WebSocket updates** on:
   - Request routing (which server handled it).
   - Skipped servers (unhealthy).
   - Health status updates.

---
