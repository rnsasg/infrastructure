## What is Amazon Route 53?
Amazon Route 53 is a Domain Name System (DNS) service that enables developers and businesses to route internet traffic to their applications. It provides a variety of features, including domain registration, health checking, traffic routing, and DNS failover, to help ensure the availability and scalability of applications.

## What are the benefits of using Amazon Router 53?

- [ ] Reliability : highly available and reliable, with multiple layers of redundancy 
- [ ] Scalability : handle millions of requests per second,suitable for applications of any size
- [ ] Cost-effectiveness: pay-as-you-go 

- [ ] Flexibility:  routing policies, including simple round-robin routing, failover routing, and geolocation routing,
- [ ] Integration with other AWS services:
- [ ] Domain registration
- [ ] Advanced features


## What are three services available on Route 53?

- [ ] Domain Name Registration:
- [ ] Domain Name System (DNS) Management
- [ ] Health Checking

## Who can we use Route 53 to route users? geographical location | routing policy | health of your application

You can use Amazon Route 53 to route users to any internet application, including web applications, mobile applications, and APIs. Route 53 enables you to route traffic based on a variety of factors, such as the geographical location of the user, the health of your application, and the routing policy that you specify.

## What are the actions performed by Route 53?

Domain name registration
DNS record management
Traffic routing
Health checking
DNS failover

## 14.How do Amazon Route 53 Resolver 53 DNS Firewall and AWS Network Firewall differ in protection against malicious DNS query threats?

Route 53 Resolver 53 DNS Firewall  ---> Outbound malicious query 
AWS Network Firewall ---> Outbound malicious query 

## How can we add a load balancer to Route 53? 

* Create a load balancer:  Application Load Balancer, Network Load Balancer, and Classic Load Balancer. 
* Configure the load balancer :  listeners, security groups, and target groups.
* Create a hosted zone
* Create a record set
* Test your configuration

##  What is the default TTL setting for records created in Amazon Route 53?
The default TTL setting for records created in Amazon Route 53 is 1 hour.

## Is it possible to route traffic based on user location using Amazon Route 53? If yes, then how?

--> by creating a geolocation resource record set

## What is the default limit of the domain supported by Route 53? 

The default limit for the number of domain names that you can register with Amazon Route 53 is 50 per AWS account

## 37. Difference between Route 53 aliases and CNAME?

## Here is a comparison of Amazon Route 53 aliases and CNAME records:

| Feature                       | Alias Record                                                                 | CNAME Record                                                             |
|-------------------------------|------------------------------------------------------------------------------|--------------------------------------------------------------------------|
| **Purpose**                   | To map a domain name or subdomain to another AWS resource or to an external resource | To map a domain name or subdomain to another domain name                 |
| **Supported resource types**  | AWS resources (e.g., ELB, CloudFront distributions, S3 buckets) and external resources (e.g., non-AWS resources with a public IP address or a domain name) | Any domain name                                                          |
| **Cost**                      | No additional charge                                                         | Standard Route 53 charges apply                                          |
| **Availability**              | Highly available                                                             | Depends on the availability of the target domain                         |
| **Latency**                   | Low                                                                          | Depends on the latency of the target domain                              |
| **DNS propagation time**      | Low                                                                          | Depends on the TTL of the target domain                                  |
| **Support for weighted routing** | Yes                                                                        | No                                                                       |
| **Support for failover routing** | Yes                                                                        | No                                                                       |

## What are the different Route53 routing policies?

Simple routing
Weighted routing
Latency-based routing
Failover routing
Geolocation routing
Multivalue answer routing

Recordset 


## What happens if all of my endpoints are unhealthy?
Route 53 can only fail over to an endpoint that is healthy. If there are no healthy endpoints remaining in a resource record set, Route 53 will behave as if all health checks are passing.

