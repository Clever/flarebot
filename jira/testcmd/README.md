# jira-cli

To more easily test the jira subpackage, this directory creates an executable called `jira-cli` that you can use to access some of the jira functions directly.

## usage

First set the following environment variables:
- **JIRA_ORIGIN**
- **JIRA_PROJECT_ID**
- **JIRA_PRIORITIES**
- **JIRA_ISSUETYPE_ID**
- **JIRA_USERNAME**
- **JIRA_PASSWORD**

```
jira-cli --help
Usage:
  jira-cli [command]

Available Commands:
  getTicket      print the captured jira ticket attributes
  getUserByEmail print the user record for a jira user
  help           Help about any command
  setDescription replace the description with the text

Use "jira-cli [command] --help" for more information about a command.
```
