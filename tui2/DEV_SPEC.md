DevSpec Document
=================

### Table of Contents
*   [General Architecture](#general-architecture)
*   [API Endpoints](#api-endpoints)
*   [Authentication and Authorization](#authentication-and-authization)
    *   [User Accounts](#user-accounts)
    *   [Role-Based Access Control](#role-based-access-control)
*   [Abstraction Layer, API Key Management, and Model Configuration](#abstraction-layer-api-key-management-and-model-configuration)
*   [Observability and Response Preparation](#observability-and-response-preparation)
*   [Performance Monitoring](#performance-monitoring)
*   [Real-Time Response Support](#real-time-response-support)

### General Architecture
-----------------------

The system is designed to handle multiple interactions with local applications utilizing device-specific sensors for home automation tasks. It should allow seamless integration of device capabilities without worrying about low-level details.

This is achieved by providing a custom abstract SDK, enabling users to easily integrate LLM abilities into their own code.

### API Endpoints
----------------

Each supported interaction is represented by an associated REST API endpoint that can be accessed directly from the application's codebase. JSON Web APIs will serve both mobile and web platforms with a focus on scalability.

The system emphasizes simple scaling, making it suitable for smaller to medium-sized applications or enterprises looking for cost-efficiency without sacrificing robustness.

### Authentication and Authorization
---------------------------------

#### User Accounts

*   Users can authenticate using local database storage and a role-based permission system.
*   Each user has access to multiple roles (see below), and their corresponding credentials are stored in an associated API key. The OAuth system requires manual admin intervention for modifying passwords.

Roles:
| Role        | Description                    |
|-------------|--------------------------------|
| Admin       | Full control over the application and data |
| Standard User | Limited functionality, mostly limited to usage |
| Guest User | Simple access based on specific data |

#### Role-Based Access Control

A user's role-based permissions determine their ability to perform actions within the system.

### Abstraction Layer, API Key Management, and Model Configuration
-----------------------------------------------------------------

#### Custom Abstract SDK

Design a custom abstract SDK that abstracts away model retrieval process, allowing applications seamlessly integrate LLL capabilities without worrying about low-level details. The model retrieval should provide users with flexibility in selecting models suitable for handling various scenarios based on device specifications.

#### Centralized Repository

Each user or project has its associated directory for storing API keys, base URLs, and model files within a centralized repository.

#### Model Settings Handling

A dedicated interface will be provided to configure the settings options including temperature sets, prompt libraries, etc. from the user-friendly view.

### Observability and Response Preparation
---------------------------------------------

#### Performance Monitoring

Provide APIs to track with cloud-based monitoring services such as AWS X-Ray for enhanced tracking. Real-time metrics regarding each endpoint are made available through this feature.

*   **Implementation Plan:**

    *   First step, set up AWS X-Ray support on the gateway's backend
    *   After successful implementation of X-Ray,
        +  Second step is connecting API and adding related end-to-end tracking capabilities through a monitoring GUI

#### Real-Time Response Support

The system includes a feature to support real-time responses. Optionally, queuing systems can be implemented for handling incoming requests.

*   **Implementation Plan:**

    *   Integrate Redis-based caching for enhanced response times
        +  First step is selecting the best suitable LRU policy cache interval & size 
    *   Implement automated tools for performance monitoring, handling latency, or metrics tracking with cloud monitoring systems

### Testing Plan
--------------

Testing follows best-practices and comprehensive testing methodologies. Tests include unit tests, end-to-end integration testing to ensure seamless integration between different components within the system.

Unit tests cover every functionality (functionalty & method) at an individual level ensuring robustness of code snippet or component.

Integration testing verifies interactions between different sub-modules for proper working while taking care of test isolation from other modules for consistency.

### Branching Strategy
---------------------

The system uses GitFlow branching strategy to manage branches efficiently. A common branch strategy that allows developers work on new features without affecting production, while still maintaining compatibility and version control with each release.

Developers can maintain a healthy workflow by working on individual feature branches until they are stable. This is supplemented by regular code reviews before it is merged into the main branches (development & staging branches).

The final merge into the master will occur at most once per month to maintain minimal conflict while maintaining production stability and progress towards release goals.



### Security Considerations
-----------------------------

Following a strong stance on security considerations using best practices available in our current implementation ensures a stable secure environment with least data breaches in case any unauthorized attacks are performed.

The focus is always on minimizing potential risks associated with the application's performance, especially related to concurrent connections and high-traffic APIs. This includes implementing robust encryption protocols throughout client-server communication, proper error handling mechanisms among others



### Performance Bottleneck Minimization
-----------------------------------------

In a multi-threaded setup, shared resources between different threads can be difficult to manage. A synchronization mechanism (either mutexes/semaphores) will be implemented around the shared resource to ensure thread safety.


Our system aims for optimizing performance through using fixed-size LRU caching mechanisms to manage model updates efficiently.
A distributed locking system is not necessary given the scope of our implementation



### DevOps and Infrastructure
-----------------------------

DevOps infrastructure includes using Redis-based caching mechanism combined with time-evolving least-recently-used (LRU) policy for optimal response times.


This approach significantly simplifies both caching and updating functionalities to improve system performance and minimize response latency under the expected load condition.
The centralized lock manager will ensure the integrity of API endpoint data despite concurrent access from different devices, minimizing conflicts that could lead to deadlocks or performance degradation.

### Committing Strategy
---------------------

Commits on master branch happen once every month. The main development workflow involves feature completion, followed by code review before a potential merge into the main line (development and staging branches).