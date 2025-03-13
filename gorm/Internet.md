---
title: Internet
created: '2025-01-10T08:07:27.153Z'
modified: '2025-01-11T02:32:56.554Z'
---

# Internet
the internet is a network of network.
## How does the internet work?
The Internet is a global network of interconnected computers that communicate using standardized protocols, primarily TCP/IP. When you request webpack, your device send a data packet through ypur internet service provider(ISP) to a DNS server, which translates the website's domain name into an IP adress. The packet is then routed across various network(using routers and swiches) to the destination server, which processes the request and sends back the response. This  back-and-forth exchange  enables the transfer of data like web packs, emails, and files, making thr internet a dynamic, decentralized system for global communication.

At a high level, the internet works by connecting devices and computer systems together using a set of standardized protocols. These protocols define how information is exchanged between devices and ensures that data is transmitted reliably and securely.

The core of the internet is a global network of interconnected routers, which are responsible for directing traffic between different devices and systems. When you send data over the internet, it is broken up into small packets that are sent from your device to a router. The router examines the packet and forward it to the next router in the path towards its destination. This process continues until the packet reaches its final destination.

To ensure that packets are sent and received correctly, the internet uses a variety of protocols, including the Internet Protocol(IP) and the Transmission Control Protocol(TCP). IP is responsible for routing packets to their correct destination, while TCP ensures that packets are transmitted reliably and in the correct order.

In addition to these core protocols, there are a range of other technologies and protocol that are used to enable communication and data exchange over the internet, including the Domain Name Sstem(DNS), the Hypertext Transfer Protocol(HTTP), and the Secure Sockets Layer/Transport Layer Security (SSL/TLS) protocol. As a developer, it is important to have a solid understanding of how these differnt technologies and protocols work together to enable communication and data exchange over the internet.

## Basic Concepts and Terminology
+ __Packet__: A small unit of data that is transmitted over the internet.
+ __Router__: A device that directs packets of data between different networks.
+ __IP Address__： A unique identifier assigned to each device on a network, used to route data  to correct destination.
+ __Domain Name__: A human-readable name that is used to identify a website, such as google.com
+ __ DNS__: The Domain Name System is responsilbe for translating domain names into IP address.
+ __HTTP__: The Hypertext Transfer Protocol is used to transfer data between a client(such as a web browser) and a server(such as a website).
+ __HTTPS__: An encrypted version of HTTP that is used to provide secure communication between a client and server.
+ __SSL/TLS__: The Secure Sockets Layer and Transport Layer Security protocols are used to provide secure communication over the internet.

## The Role of Procotols in Internet
Protocols play a critical role in enabling communication and data exchange over the internet. A protocol is a set of rules and standards that define how information is exchanged between devices and systems.

There are many different protocols used in internet communication, including the Internet Protocol(IP), the Transmission Control Protocol(TCP), the User Datagram Protocol(UDP), the Domain Name System(DNS), and many others.

IP is responsible for routing packets of data to their correct destination, while TCP and UDP ensure that packets are transmitted realiably and efficiently. DNS is used to translate domain names into IP address, and HTTP is used to transfer data between clients and server.

One of the key benefits of using standardlized protocols is that they allow devices and systems from different manufacturers and vendors to communicate with each other seamlessly. For example, a web browser developed by one company can communicate with a web server developed by  another company, as long as they both adhere to the HTTP protocol.

## Understanding IP Address and Domain Names
IP addresses and Domain names are both important concepts to understand when working with the internet.

An IP address is a unique identifier assigned to each device on a network. It's used to route data to the correct destination, ensuring that information is sent to the intended recipient. IP address are typically represented as a series of four numbers separated by period, such as "192.168.1.1".

Domain names, on the other hand, are human-readable names used to identify websites and other internet resources. The're typically composed of two or more parts, separated by periods. For example, "google.com" is a domain name. Domain names are translated into IP address using the Domain Name System(DNS).

DNS is a critical part of the internet infrastructure, responsible for translating domain names into IP address. When you enter a domain name into your brower, your computer sends a DNS query to a DNS server, which returns the corresponding IP address. Your computer then uses that IP address to connect to the websites or other resource you've requested.

## Introduction to HTTP and HTTPS
HTTP(Hypertext Transfer Protocol) and HTTPS(HTTP Secure) are two of the most commonly used protocols in internet-based applications and services.

HTTP is the protocol used to transfer  data between a client(such as a web browser) and a server (such as a website). When you visit a website, your web browser sends an HTTP resource you've requested, asking for the webpage or other resourece you've requested. The server then sends an HTTP response back to the client, containing the requested the requested data.

HTTPS is a more secure version of HTTP, which encrypts the data being transmitted between the client and server using SSL/TLS(Secure Sockets Layer/Transport Layer Security) encryption. This provides an additional layer of security, helping to protect sensitive information such as login credentials, payment information, and other personal data.

When you visit a website that uses HTTPS, your web browsers will display a padlock icon in the address bar, indicating that the connection is secure. You may also see the letters"https" at the begining of the website address, rather than "http".

## Buliding Application with TCP/IP
TCP/IP(Transmission Control Protocol/Internet Protocol) is the underlying communication protocol used by most internet-based applications and services. It provides a reliable, ordered, and error-checked delivery of data between applications running on different devices.

When building applications with TCP/IP, there are a few key concepts to understand:
  + __Ports:__ Ports are used to identify the application or service running on a device. Each application or service is assigned a unique port number, allow data to be sent to the correct destination.

  + __Socket:__ A soccket is a combination of an IP address and a port number, representing a specific endpoint for communication. Sockets are used to establish connections between devices and transfer data between applications.

  + __Connections:__ A connection is established between two sockets when two devices want to communicate with each other. During the connection establishment process, the devices negotiate various parameter such as the maximum segment size and window size, which determine how how data will be transmitted over the connection.

  + __Data transfer:__ Once a connection is established, data can be transferred between the applications running on each device. Data is typically transmitted in segments, with each segment containing a sequence number and other metadata to ensure reliable delivery.

When building applications with TCP/IP, you'll need to ensure that your application is designed to work with the appriate ports, sockets, and connections. You'll also need to familiar with the various protocols and standards that are commonly used with TCP/IP, such as HTTP,FTP(File Transfer Protocol), and SMTP(Simple Mail Transfer Protocol). Understanding these concepts and protocols is esstial for building effective, scalable, and secure internet-based applications and services.

## Securing Internet Communication with SSL/TLS
As we discussed earlier, SSL/TLS(Secury Socket Layer/ Transmission Layer Secury) is a protocol used to encrypt data being transmitted over the internet. It is commonly used to provide secure connections for applications such as web browsers,email clients, and file transfer programs.

When using SSL/TLS to secure internet communication,there are a few key concepts to understand:
  + __Certificates:__ SSL/TLS certificates are used to established trust between the client and server. They contain information about the identity of the server and signed by a trusted third party(a Certificate Authority) to verify their authenticity.

  + __Handshake:__ During the SSL/TLS handshake process, the client and server exchange information to negotiate the encreption algorithm and other parameters for the secure connection.

  + __Encryption:__ Once the secure connection is established, data is encrypted using the agreed-upon algorithm and can be transmitted securely between the client and server.

When building internet-based applications and services, it's important to understand how SSL/TLS works and to ensure that your application is designed to use  SSL/TLS when transmitting sensitive data such as login credentials, payment information, and other personal data. You'll also need to ensure that you obtain and maintain valid SSL/TSL certificates for your servers, and that you follow best practices for configuring and securing your SSL/TLS connections. By doing so, you can help protect your user's data and ensure the integrity and confidentially of your application's communication over the internet.

## Conclusion
And that brings us to the end of this article. We've covered a lot of groud, so let's take a moment to review what we've learned:
+ The internet is a global network of interconnected computers that uses a standard set of communication protocols to exchange data.

+ The internet works by connecting devices and computer systems together using standaraized protocols,such as IP and TCP.

+ The core of the internet is a global network of interconnected routers that direct traffic between different devices and systems.

+ Basic concepts and terminology that you need to familiarize yourself with include packets, routers, IP addresses, domain names, DNS, HTTP, HTTPS, and SSL/TLS.

+ Protocols play a critical role in enabling communication and data exchange over the internet, allowing devices and system from different manufactuers and vendors too communicate seamlessly.






