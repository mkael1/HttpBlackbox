## Why I Built This

Over the past few years, I've been working extensively with web technologies, but I've realized that I never really took the time to understand how HTTP actually works under the hood. Most of the web development work we do relies on frameworks and libraries, which are incredibly useful—but they can also feel like black boxes.  

This year, I set a personal goal to uncover some of those black boxes and deepen my understanding of the fundamentals. What better way to learn than to implement an HTTP server myself, from scratch, on top of raw TCP in Go?  

This project is my exploration into the inner workings of HTTP, TCP connections, and the lower-level code that power's the web. It's been challenging, but incredibly rewarding


## Features

- **Accept TCP connections** – Handles multiple incoming client connections.
- **Parse HTTP requests** – Reads and interprets request lines, headers, and bodies.
- **Custom handlers** – Allows defining your own handlers to respond with dynamic content.
- **Serve files** – Easily respond with static files.
- **Chunked responses** – Support for sending responses in chunks when needed.
- **Serve HTML** – Respond with HTML content directly.
