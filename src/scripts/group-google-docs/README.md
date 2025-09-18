# Group google docs

idempotent script to move flare docs from flarebot service account root level into shared drive.


## SETUP

`export CONF=$(ark secrets read development.flarebot google-flarebot-service-account-conf)`
`export JIRA_PASSWORD=$(ark secrets read development.flarebot jira-password)`
