<div align="center">
  <img width="128" src="./logo.png" alt="Okta logo" />
  <h1>Okta Auth Plugin for OpenVPN</h1>
  <p>Plugin for authenticating OpenVPN users against Okta optionally using MFA</p>
  <hr />
  <br />
  <a href="#">
    <img src="https://img.shields.io/badge/stability-alpha-ff69b4?style=for-the-badge" />
  </a>
  <a href="https://en.wikipedia.org/wiki/MIT_License" target="_blank">
    <img src="https://img.shields.io/badge/license-MIT-maroon?style=for-the-badge" />
  </a>
  <a href="#">
    <img src="https://img.shields.io/badge/support-community-purple?style=for-the-badge" />
  </a>
  <a href="https://conventionalcommits.org" target="_blank">
    <img src="https://img.shields.io/badge/Conventional%20Commits-1.0.0-orange.svg?style=for-the-badge" />
  </a>
</div>
<br />
<hr />
<br />

- [👁️ Overview](#️-overview)
- [✅ Requirements](#-requirements)
- [⛏️ Build Process](#️-build-process)
- [⚙️ Configuration](#️-configuration)
- [🔑 Logging into OpenVPN](#-logging-into-openvpn)
- [🔗 Additional Information](#-additional-information)
- [📃 License](#-license)
- [❓ Questions, Issues and Feature Requests](#-questions-issues-and-feature-requests)

## 👁️ Overview

The Okta OpenVPN plugin allows you to authenticate your OpenVPN users using Okta's SSO and MFA technology using pure REST API calls. While Okta's official support is thorugh the use of RADIUS, this plugin does not require the use of a RADIUS server. The plugin has been tested with OpenVPN Community Edition, but it should work with Access Server as well.

## ✅ Requirements

- Setup [Okta SSO and MFA](https://okta.com) for your organization.
- Review the documentation for [OpenVPN Server](https://community.openvpn.net/openvpn) for details on how to configure OpenVPN.
- Alpine, Ubuntu or CentOS/RedHat Linux
- OpenVPN libraries and headers must be installed

## ⛏️ Build Process

To build the code from source, you will also need the following tools: `cc` or `gcc`, `make`, `go`

1. Check out the latest code from the repository to your local system and make sure your `GOPATH` is set up properly. Refer to the [Go Documentation](https://golang.org/doc/) for details.
1. Run `make` to build the binaries. You may need to run `CFLAGS=-I/usr/include/openvpn make` if your OpenVPN headers are located in a subdirectory of `/usr/include` on your OS.
1. Copy the contents of the `build` folder to your OpenVPN plugins directory along with the `configs/okta-openvpn.yml` sample configuration.

## ⚙️ Configuration

Once you have the binaries compiled and place into your OpenVPN plugins folder, you'll need to add the lines below to your OpenVPN server configuration. These instructions assume you're using `/usr/lib/openvpn/plugins/okta-openvpn` for storing the plugin files.

```openvpn.conf
plugin /usr/lib/openvpn/plugins/okta-openvpn/auth_script.so /usr/lib/openvpn/plugins/okta-openvpn/okta-openvpn -conf /usr/lib/openvpn/plugins/okta-openvpn/okta-openvpn.yml
client-cert-not-required
username-as-common-name
```

Edit the `okta-openvpn.yml` file and modify settings according to your organization and needs. You **must** supply a value for `org_name`, which is typically your Okta SSO hostname without the `.okta.com` suffix. The remainder of the settings are explained within the sample file and are optional.

If you choose to use an API key, you'll need to follow one of the following articles depending on your Okta subscription:

- <https://developer.okta.com/docs/api/getting_started/getting_a_token>
- <https://support.okta.com/help/s/article/How-do-I-create-an-API-token>

If you wish to add location details to the log output in addition to just the client IP, you'll need to download the GeoLite2 City database from <https://dev.maxmind.com/geoip/geoip2/geolite2/> and specify the path to the extracted `.mmdb` file in your configuration.

Finally, you'll need to make sure the `auth-user-pass` directive is specified in your OpenVPN client configuration so that clients are prompted for a username and password.

## 🔑 Logging into OpenVPN

Once the plugin has been configured on your OpenVPN server, users can log in using their Okta credentials. If their account is protected using MFA, they have 2 choices on how to supply the additional factor of authentication:

1. If the `totp` method is enabled in the configuration file, users can append a `+` sign to their password followed by the 6 digit code from their Okta Verify, Google Authenticator, etc. mobile app. In this case, users will not be able to save their OpenVPN credentials as their password will change each time since the 6 digit OTP code changes regularly.
1. If the `push` method is enabled in the configuration file, users can simply enter their password by itself. A push request will be sent automatically to the Okta Verify mobile app. Users have a given amount of time, which is configurable in the `okta-openvpn.yml` file, to respond to the push request before it times out.

## 🔗 Additional Information

- [OpenVPN Server](https://community.openvpn.net/openvpn)
- [Okta](https://okta.com)
- [Okta Developer API Tokens](https://developer.okta.com/docs/api/getting_started/getting_a_token)
- [Okta SSO API Tokens](https://support.okta.com/help/s/article/How-do-I-create-an-API-token)
- [Go Documentation](https://golang.org/doc/)

## 📃 License

This module is distributed under the MIT License.

## ❓ Questions, Issues and Feature Requests

If you have questions about this project, find a bug or wish to submit a feature request, please [submit an issue](https://github.com/josh-hogle/okta-openvpn-auth-plugin/issues).
