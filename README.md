# Bitbucket Cascade Merge

Bitbucket Cascade Merge is a service used to automatically cascade changes
after a pull request has been successfully merged (fulfilled). This feature
has not been ported to Bitbucket Cloud.

More information can be found on the Bitbucket Server
[automatic branch merging](https://confluence.atlassian.com/bitbucketserver/automatic-branch-merging-776639993.html)
documentation. Besides semantic version that is currently not supported, the behaviour tries to be the same.

You can show you interest and vote for this feature :
[BCLOUD-14286](https://jira.atlassian.com/browse/BCLOUD-14286)

## Usage

### What do you need ?

* A service to host this service (eg. Amazon ECS)
* An account on bitbucket.org with r/w privileges to the project

### Get an API Token

The API token will be used to call the Bitbucket API and fetch/push the
repository.

1. Open https://id.atlassian.com/manage-profile/security/api-tokens
2. Click on **Create API token with scopes** and select **Bitbucket**
3. Type a **label**, pick an **expiry** and select the following scopes :
   * `read:repository:bitbucket`, `write:repository:bitbucket`
   * `read:pullrequest:bitbucket`, `write:pullrequest:bitbucket`
4. Copy the token somewhere safe, you will need it later to configure
   environment variables

App passwords used to fill this role, but Atlassian stops accepting them on
2026-06-09 and removes them on 2026-07-28. Tokens expire (one year at most),
so plan to rotate `BITBUCKET_PASSWORD` before yours lapses.

### Configure a webhook on the repository

1. Navigate to the repository you want to activate cascade merges
2. Go to Settings > Workflow > Webhooks
3. Click on **Add webhook**
4. Type a title, the url of your container and select
   **Choose from a full list of triggers** : Pull Request > Merged

Security notice: you can use a *token* query parameter in the url field
(eg. `?token=your-random-token`) that needs to match the configured value
of the `TOKEN` environment variable.

### Configure the container

The container can be configured with environment variable.

| Key                | Default Value | Description                                                   |
|--------------------|---------------|---------------------------------------------------------------|
| PORT               | 5000          | Server will listen on this port                               |
| BITBUCKET_USERNAME |               | Atlassian account email that owns the API token               |
| BITBUCKET_PASSWORD |               | Bitbucket API token                                           |
| TOKEN              |               | Security token                                                |

`BITBUCKET_USERNAME` must be the email address of the Atlassian account
that created the token (e.g. a service account like `ci@example.com`), not
`x-bitbucket-api-token-auth`. Scoped API tokens are rejected with a 401
(`API token must be used with an atlassian registered email`) unless the
username is that account's email.





### Run the container

Was initially created by [Samuel Contesse](https://github.com/samcontesse).


```
docker run \
  -e BITBUCKET_USERNAME=you@example.com -e BITBUCKET_PASSWORD=<fillme> -e TOKEN=<fillme> \
  --publish 5000:5000 \
  --name bcm \
  morpheancloud/bitbucket-cascade-merge
```

## Requirements

[Libgit2 v1.5.1](https://github.com/libgit2/libgit2/archive/refs/tags/v1.5.1.tar.gz)
must be installed if you do not use the Docker image provided.

