AGENTS.md
=================
Local LLM Interaction AI-Gateway
=====================================

Problem Statement and Target Audience
--------------------------------------

### Application Type

*   Local software developers who will develop or run applications that utilize device-specific sensors for home automation tasks.

### Ideal User

*   Primary developers of local applications.
*   Users who need to maintain provider abstraction and simplify local AI development.

Gateway's Role and Platforms
------------------------------

### Interaction with Local Applications

Implement REST API endpoint for each supported interaction (e.g., language processing, entity recognition) allowing applications to call these endpoints directly from within their own code.

### Supported Platforms

JSON Web APIs for mobile and web platforms. Focus on simple scaling.

Authentication and Authorization
---------------------------------

*   Define user accounts with a manual admin intervention OAuth system.
*   Implement role-based access control (RBAC): define multiple roles and corresponding API keys, with options to create guest users.
*   Consider adding an audit log feature to track changes to users' roles, permissions, and account credentials.

Abstraction Layer, API Key Management, and Model Configuration
----------------------------------------------------------------

### Custom Abstract SDK

Design a custom abstract SDK that abstracts away the model retrieval process, allowing applications to seamlessly integrate LLM capabilities without worrying about low-level details.

Integrate with Redis-based caching for optimal response times. Consider using AWS X-Ray for cloud-based performance monitoring.

### Centralized Repository

Create separate directories for each user or project in the Gateway's file system, where API keys, base URLs, and model files are stored and easily accessible through a RESTful interface.

Implement Model Settings Handling that allows users to configure settings options from a user-friendly interface.

### Observability and Response Preparation

Implement performance monitoring with AWS X-Ray. Consider implementing Redis-based caching for optimal response times.

Create APIs for tracking with cloud-based monitoring services like AWS X-Ray, in addition to providing automated tools for handling latency and promoting metrics related to API request-response sequences.

Implement automated tools for monitoring the gateway's performance on startups, shutdowns, and during any time of active usage.

### Testing Plan

*   Implement unit tests for each functionality at an individual level.
*   Develop comprehensive integration testing covering interactions between different submodules.
*   Conduct code reviews regularly to ensure a high-quality codebase and consistency throughout.

Consider adopting the GitFlow branching strategy for managing branches efficiently.

Security Considerations
-----------------------------

Incorporate a robust security posture using best practices available in our current implementation, including encryption protocols throughout client-server communication, proper error handling mechanisms, secure authentication and authorization mechanisms, secure database practices, security monitoring tools for early detection of malicious actors, and a comprehensive backup system to protect critical data.

Implementation
----------------

Design the Gateway's architecture in collaboration with local developers who have deep expertise in creating intuitive interfaces. Collaborate with AI modelists and hardware specialists to enable better integration of device capabilities for seamless interactions and development workflows.


Consider integrating our solution with an existing IAC (Infrastructure-as-Code) environment that uses tools like Terraform or AWS CloudFormation, such that the implementation needs minimal additional work.

Committing Strategy
-----------------

Follow best practices, but use regular, controlled releases to ensure data integrity throughout this continuous integration process.