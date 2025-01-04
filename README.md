
# Apee-i

Command Line based api-tester utility completely written in [Golang](https://go.dev)


[![MIT License](https://img.shields.io/badge/License-MIT-green.svg)](https://choosealicense.com/licenses/mit/)
[![Golang](https://img.shields.io/badge/Language-Golang-blue.svg)](https://go.dev)


## About

Apee-i is a CLI-Based API Tester, a command-line tool written in Go for testing APIs efficiently. It uses a JSON file to define API requests and allows you to execute individual requests or custom pipelines of multiple API calls. The tool is designed to be lightweight, flexible, and developer-friendly, making it an essential utility for testing APIs in development environments.

*NOTE*: Currently only supports when API responds with JSON structure
## Features

- JSON-based Configuration: Define API requests and pipelines in a single JSON file.

- Single Request Execution: Run individual API tests with ease.

- Custom Pipelines: Chain multiple API requests to test workflows and dependencies.

- Detailed Logging: Provides detailed output for each request, including status codes, response bodies, and errors.

- Cross-Platform: Compatible with major operating systems (Linux, macOS, Windows).
## Installation

### By Cloning:

Have `Go` installed in your system

Step 1. Clone
```
git clone https://github.com/IbraheemHaseeb7/apee-i.git
cd apee-i
```

Step 2. Install using go module handler
```
go install
```



## Usage

### Step 1. Create a json file
You can name your json file `api.json` if you don't want to mention the filename everytime when you hit the command or otherwise you will have to tell the tool using `--file` flag like so
```
apee-i --file=myfile.json
```

### Follow the given json file structure
`api.json` put this file wherever you want to test your APIs. A sample file is present [here](https://github.com/IbraheemHaseeb7/apee-i/blob/main/example/api.json) for you to use.


```json
{
	"baseUrl": {
		"development": "http://localhost:8000/api",
		"staging": "http://staging.com/api",
		"production": "http://production.com/api"
	},
	"credentials": {
		"development": {
			"email": "example@gmail.com",
			"password": "Example@123"
		},
		"staging": {
			"email": "example@gmail.com",
			"password": "Example@123"
		},
		"production": {
			"email": "example@gmail.com",
			"password": "Example@123"
		}
	},
	"loginDetails": {
		"route": "/login",
		"type": "JWT",
		"token_location": "data.access_token"
	},
	"current_pipeline": [
		{ "endpoint": "/test" },
		{
			"endpoint": "/test",
			"method": "POST",
			"body":  {
                "name": "John Doe",
                "email": "johndoe@gmail.com",
			},
			"expectedStatusCode": 201,
			"headers": {
				"X-HEADER": "SOME_VALUE"
			}
		}
	],
	"custom_pipelines": {
		"users": [
			{ "endpoint": "/users" },
			{ 
                "endpoint": "/users/1",
                "method": "PATCH",
                "body": {
                    "name": "Sara Doe"
                }
            }
		],
		"test": [
			{
				"endpoint": "/test"
			}
		]
	}
}
```

## How to use now?

### Login Details

1. Provide the login route for the apee-i to hit
2. Currently it only supports JWT auth
3. Token location is the field where the access token will be available in JSON response

### Select environement and credentials by

*NOTE*: Default environement is `development` if you dont provide with the flag

```
apee-i --env=staging
```

### Select file by

*NOTE*: Default filename is `api.json` if you dont provide with the flag

```
apee-i --file=myfile.json
```

### Select pipelines by

*NOTE*: Default pipeline is `current` if you dont provide with the flag

```
apee-i --pipeline=current
```
This executes all the endpoints in `current_pipline` field

OR 

You can run a selected custom pipeline by
```
apee-i --pipeline=custom --name=users
```
This executes all the endpoints in `custom_pipelines` under `users`

OR 

Finally you can run all the pipelines in the `custom_pipelines` by
```
apee-i --pipeline=custom --name=all
```
## 🔗 Find me here
[![portfolio](https://img.shields.io/badge/my_portfolio-000?style=for-the-badge&logo=ko-fi&logoColor=white)](https://ibraheemh.vercel.app/)
[![linkedin](https://img.shields.io/badge/linkedin-0A66C2?style=for-the-badge&logo=linkedin&logoColor=white)](https://www.linkedin.com/in/ibraheemhaseeb7)


## Authors

- [@IbraheemHaseeb7](https://www.github.com/IbraheemHaseeb7)

