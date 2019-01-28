# cane-project

Cisco API Normalization Engine - Unify & Normalize API's across multiple Cisco platforms.


## Business/Technical Challenge

With IT teams turning more and more to API’s to build, configure, and maintain their complex infrastructure, the explosion of API standards, formats, and authentication methods have made consuming them in a holistic fashion extremely difficult. Despite the fact that API’s were created to help us automate repetitive tasks that can be easily codified, one of the biggest hurdles for API’s is that there isn't a true standard.  Every vendor has their own, and sometimes several distinct API’s and further complicate things, the way that you authenticate to them, even within a single vendors API set, could be completely different. All of these caveats undermine the goal of automation - causing more work, not less - as was the original intent.

## Proposed Solution

Introducing CANE - an API Aggregation platform that can consume multiple underlying Cisco (and other vendor) API’s, and allow you to build your own business-centric API’s that make sense for your organization. CANE promotes programmability and automation via a single, vendor agnostic platform.  Using a composable API engine, our customer can create a mapping between multiple underlying platform API’s (e.g. NX-API), and chain them together to create an outcome, not just an API response.

Authentication:

In order to maintain security, many API’s use some sort of authentication to protect access to sensitive information on the device or for changing the device configuration.  However, this security can also be an unreasonable hindrance to the API's usefulness...the frequency of re-authentications, distribution of credentials, etc.

CANE can act as both the authentication provider and authentication subscriber. As an authentication provider, each user can securely authenticate into CANE with a username and password (additional authentication methods will be incorporated in the future such as SSO).  After this initial authentication, the user will be issued a Java Web Token (JWT) that CANE will maintain. This token then allows the user/device to make subsequent API calls into CANE without manual re-authentication, or having to store a large number of different device credentials locally.

As an authentication subscriber, in the background, CANE will also maintain active Authentications into each vendor device, renewing session tokens automatically as needed.  This ensures that the endpoint API is always available for consumption - the first time and every time.  If the API requires an API key, CANE will centrally maintain all of your API keys on your behalf.  This ensures that there is a single, protected holder of the API key.  When a user authenticates into CANE with their username and password, they are given the right to use the API key while never having it within their possession.  If the API key needs to be changed due to a periodic refresh or a compromise of the key, no problem.

Business-centric API’s:

The reality is that we rarely just call a SINGLE API. Multiple API calls are require to complete a business process, whether that is turning up a new branch location, or provisioning a new network across multiple different switch & routing platforms. CANE is centered around the idea of a workflow. Workflows allow you to chain multiple disparate API’s calls together, mapping outputs of some API’s, to inputs of others. Composing unique business API’s in this fashion allows you consume your infrastructure in a much more effective way, and provide your company API’s that deliver on YOUR needs.


### Cisco Products Technologies/ Services

Our solution will complement all Cisco technologies that leverage RESTful APIs, including (but not limited to):

* [Application Centric Infrastructure (ACI)](http://cisco.com/go/aci)
* [DNA Center (DNA-C)](http://cisco.com/go/dna)
* [UCS Director](http://cisco.com/go/ucsdirector)
* [Firepower Threat Defense](http://cisco.com/go/ngfw)
* [Identity Services Engine](http://cisco.com/go/ise)
* [Catalyst 9000](http://cisco.com/go/catalyst)
* [Nexus 9000](http://cisco.com/go/nexus)
* [Meraki](http://meraki.cisco.com)

## Team Members

* Matthew Garrett <matgarre@cisco.com> - Commercial
* Kevin Redmon <kredmon@cisco.com> - Commerical
* Konrad Reszka <kreszka@cisco.com> - Public Sector

## Solution Components

CANE is written in GO, and uses MongoDB as the underlying document store. The graphical interface uses Angular, which simply calls the CANE API’s on the backend.
We wanted something that could be compiled to run on various platforms, and packed in a container, VM, or whatever you’d like.

## Usage

Please refer to this video:


## Installation

Start a MongoDB Container:

sudo docker run -d --name mongodb -p 27017:27017 -v ~/data:/data/db mongo

Run the GO executable:

go run main.go

CANE is reachable (by default) on port 8005


## Documentation

(In Progress)


## License

Provided under Cisco Sample Code License, for details see [LICENSE](./LICENSE.md)

## Code of Conduct

Our code of conduct is available [here](./CODE_OF_CONDUCT.md)

## Contributing

See our contributing guidelines [here](./CONTRIBUTING.md)
