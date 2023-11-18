# Git Stats Demo

## Objective
In this demonstration, we will explore statistics on Git. We will identify the Lines Of Code (LOC) in a repository, finding all code lines. Additionally, we will observe the number of additions and deletions made by each developer.

```bash
git clone https://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@github.com/ahmetk3436/git-stats-golang
```

## Navigate to the Directory
```bash
cd git-stats-golang
```
## Docker build for CLI
```bash
docker build -t app .
```
## Run Docker Compose
```bash
docker compose up -d
```

## Make Changes

If you make any changes, you can rebuild by running the following command:

```bash
docker compose up --build -d
```
This Docker Compose file will deploy a backend, Prometheus, and Grafana server. These systems are automatically interconnected. The /metrics endpoint in the backend contains Prometheus metric outputs. Additionally, under the yaml folder, a dashboard tracking Prometheus, Grafana, and Golang metrics will be available.

## Explore the Data

Feel free to navigate to port 80 in your current environment to explore the data of my repositories. Enjoy!

## Custom API Endpoints

Here are some custom API endpoints available in the application:

## GitHub Endpoints

1. **Get All Repositories:**
    - **Endpoint:** `/api/github/repos`
    - **Description:** Retrieves all public repositories in the account with username "ahmetk3436."

2. **Get a Repository by Project ID:**
    - **Endpoint:** `/api/github/repo?projectID=<projectID>`
    - **Description:** Fetches a repository based on the specified Project ID in the GitHub API.

3. **Get All Commits in a Repository:**
    - **Endpoint:** `/api/github/commits?projectOwner=<projectOwner>&repoOwner=<repoOwner>`
    - **Description:** Gets all commits in a repository specified by the project owner and repository owner.

4. **Calculate Lines of Code (LOC) in a Repository:**
    - **Endpoint:** `/api/github/loc?repoUrl=<repoUrl>`
    - **Description:** Downloads repositories and calculates the Lines of Code (LOC).

## GitLab Endpoints

Similar endpoints are available for GitLab with the base path `/api/gitlab/`.

1. **Get All Repositories:**
    - **Endpoint:** `/api/gitlab/repos`
    - **Description:** Retrieves all public repositories in the GitLab account.

2. **Get a Repository by Project ID:**
    - **Endpoint:** `/api/gitlab/repo?projectID=<projectID>`
    - **Description:** Fetches a repository based on the specified Project ID in the GitLab API.

3. **Get All Commits in a Repository:**
    - **Endpoint:** `/api/gitlab/commits?projectOwner=<projectOwner>&repoOwner=<repoOwner>`
    - **Description:** Gets all commits in a repository specified by the project owner and repository owner.

4. **Calculate Lines of Code (LOC) in a Repository:**
    - **Endpoint:** `/api/gitlab/loc?repoUrl=<repoUrl>`
    - **Description:** Downloads repositories and calculates the Lines of Code (LOC).

Feel free to use these endpoints to interact with the GitHub and GitLab APIs.

# Example CLI Usage for Docker

Here are examples of how to use the CLI commands within a Docker container:

## GitHub Endpoints

### Get All Repositories
```bash
docker run app /bin/sh -c "./app cli --github-token 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"
```
```bash
docker run app /bin/sh -c "./app cli --github-token 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 --project-id 621058402"
```
### GitLab Endpoints
```bash
docker run app /bin/sh -c "./app cli --gitlab-host https://gitlab.youandus.net --gitlab-token glpat-FiBYym_JyJPkhsmxVydv"
```

```bash
docker run app /bin/sh -c "./app cli --gitlab-host https://gitlab.youandus.net --gitlab-token glpat-FiBYym_JyJPkhsmxVydv --project-id 3"
```
Feel free to replace the placeholders with your actual tokens, project IDs, and GitLab host URLs.

