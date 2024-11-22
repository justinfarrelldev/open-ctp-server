// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "API Support",
            "email": "justinfarrellwebdev@gmail.com"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/account/create_account": {
            "post": {
                "description": "This endpoint creates a new multiplayer account, protected by a password.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "account"
                ],
                "summary": "Create a new account",
                "parameters": [
                    {
                        "description": "account creation request body",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/account.CreateAccountArgs"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Account successfully created",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {}
                    },
                    "403": {
                        "description": "Forbidden",
                        "schema": {}
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {}
                    }
                }
            }
        },
        "/account/get_account": {
            "get": {
                "description": "This endpoint gets a multiplayer account's info.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "account"
                ],
                "summary": "Gets an account",
                "parameters": [
                    {
                        "description": "account acquisition request body",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/account.GetAccountArgs"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Account successfully retrieved",
                        "schema": {
                            "$ref": "#/definitions/account.Account"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {}
                    },
                    "403": {
                        "description": "Forbidden",
                        "schema": {}
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {}
                    }
                }
            }
        },
        "/account/update_account": {
            "put": {
                "description": "This endpoint updates an account's info.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "account"
                ],
                "summary": "Updates an account",
                "parameters": [
                    {
                        "description": "account update request body",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/account.UpdateAccountArgs"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully updated account!",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "account_id must be specified",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "an error occurred while decoding the request body: \u003cerror message\u003e",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/game/create_game": {
            "post": {
                "description": "This endpoint creates a new multiplayer game, optionally protected by a password.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "game"
                ],
                "summary": "Create a new game",
                "parameters": [
                    {
                        "description": "Game creation request body",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/game.CreateGameArgs"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Game successfully created",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {}
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {}
                    }
                }
            }
        },
        "/health": {
            "get": {
                "description": "Returns the status of the service.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "health"
                ],
                "summary": "Health check endpoint",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/health.Response"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {}
                    }
                }
            }
        },
        "/lobby/create_lobby": {
            "post": {
                "description": "This endpoint creates a new multiplayer lobby, protected by a password.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "lobby"
                ],
                "summary": "Create a new lobby",
                "parameters": [
                    {
                        "description": "lobby creation request body",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/lobby.CreateLobbyArgs"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Lobby successfully created",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {}
                    },
                    "403": {
                        "description": "Forbidden",
                        "schema": {}
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {}
                    }
                }
            }
        },
        "/lobby/get_lobby": {
            "get": {
                "description": "This endpoint gets a multiplayer lobby's info.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "lobby"
                ],
                "summary": "Gets a lobby",
                "parameters": [
                    {
                        "description": "lobby acquisition request body",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/lobby.GetLobbyArgs"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Lobby successfully retrieved",
                        "schema": {
                            "$ref": "#/definitions/lobby.Lobby"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {}
                    },
                    "403": {
                        "description": "Forbidden",
                        "schema": {}
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {}
                    }
                }
            }
        }
    },
    "definitions": {
        "account.Account": {
            "description": "Structure for representing a player account.",
            "type": "object",
            "properties": {
                "email": {
                    "description": "Email is the email address of the player.",
                    "type": "string"
                },
                "experience_level": {
                    "description": "ExperienceLevel represents the player's experience level (0=beginner, 1=easy, 2=medium, 3=hard, 4=very hard, 5=impossible)",
                    "allOf": [
                        {
                            "$ref": "#/definitions/account.ExperienceLevel"
                        }
                    ]
                },
                "info": {
                    "description": "Info contains additional information about the player.",
                    "type": "string"
                },
                "location": {
                    "description": "Location indicates the player's real-life location.",
                    "type": "string"
                },
                "name": {
                    "description": "Name is the name of the player.",
                    "type": "string"
                }
            }
        },
        "account.AccountParam": {
            "description": "Structure for representing a player account with non-required fields.",
            "type": "object",
            "properties": {
                "email": {
                    "description": "Email is the email address of the player.",
                    "type": "string"
                },
                "experience_level": {
                    "description": "ExperienceLevel represents the player's experience level (0=beginner, 1=easy, 2=medium, 3=hard, 4=very hard, 5=impossible)",
                    "allOf": [
                        {
                            "$ref": "#/definitions/account.ExperienceLevel"
                        }
                    ]
                },
                "info": {
                    "description": "Info contains additional information about the player.",
                    "type": "string"
                },
                "location": {
                    "description": "Location indicates the player's real-life location.",
                    "type": "string"
                },
                "name": {
                    "description": "Name is the name of the player.",
                    "type": "string"
                }
            }
        },
        "account.CreateAccountArgs": {
            "description": "Structure for the account creation request payload.",
            "type": "object",
            "properties": {
                "account": {
                    "description": "The account to create.",
                    "allOf": [
                        {
                            "$ref": "#/definitions/account.Account"
                        }
                    ]
                },
                "password": {
                    "description": "The password for the account to be created",
                    "type": "string"
                }
            }
        },
        "account.ExperienceLevel": {
            "type": "integer",
            "enum": [
                0,
                1,
                2,
                3,
                4,
                5
            ],
            "x-enum-varnames": [
                "Beginner",
                "Easy",
                "Medium",
                "Hard",
                "Very_Hard",
                "Impossible"
            ]
        },
        "account.GetAccountArgs": {
            "description": "Structure for the account acquisition request payload.",
            "type": "object",
            "properties": {
                "account_id": {
                    "description": "The account ID for the account that will be retrieved.",
                    "type": "integer"
                }
            }
        },
        "account.UpdateAccountArgs": {
            "description": "Structure for the account update request payload.",
            "type": "object",
            "properties": {
                "account": {
                    "description": "The account to create.",
                    "allOf": [
                        {
                            "$ref": "#/definitions/account.AccountParam"
                        }
                    ]
                },
                "account_id": {
                    "description": "The account ID for the account that will be updated.",
                    "type": "integer"
                },
                "password": {
                    "description": "The password for the account to be created",
                    "type": "string"
                }
            }
        },
        "game.CreateGameArgs": {
            "description": "Structure for the game creation request payload.",
            "type": "object",
            "properties": {
                "password": {
                    "description": "Password is the password for the game.\nThis field is required if PasswordProtected is true.\nIt must be longer than 6 characters.",
                    "type": "string"
                },
                "password_protected": {
                    "description": "PasswordProtected indicates whether the game is password-protected.\nIf true, a password must be provided.",
                    "type": "boolean"
                }
            }
        },
        "health.Response": {
            "type": "object",
            "properties": {
                "status": {
                    "type": "string",
                    "example": "OK"
                }
            }
        },
        "lobby.CreateLobbyArgs": {
            "description": "Structure for the lobby creation request payload.",
            "type": "object",
            "properties": {
                "lobby": {
                    "description": "The lobby to create.",
                    "allOf": [
                        {
                            "$ref": "#/definitions/lobby.Lobby"
                        }
                    ]
                },
                "password": {
                    "description": "The password for the lobby to be created",
                    "type": "string"
                }
            }
        },
        "lobby.GetLobbyArgs": {
            "description": "Structure for the lobby acquisition request payload.",
            "type": "object",
            "properties": {
                "lobby_id": {
                    "description": "The lobby ID for the lobby that will be retrieved.",
                    "type": "string"
                }
            }
        },
        "lobby.Lobby": {
            "description": "Structure for representing a player lobby.",
            "type": "object",
            "properties": {
                "id": {
                    "description": "ID is the unique identifier for the lobby.",
                    "type": "string"
                },
                "is_closed": {
                    "description": "IsClosed indicates if the lobby is closed.",
                    "type": "boolean"
                },
                "is_muted": {
                    "description": "IsMuted indicates if the lobby is muted.",
                    "type": "boolean"
                },
                "is_public": {
                    "description": "IsPublic indicates if the lobby is public.",
                    "type": "boolean"
                },
                "name": {
                    "description": "Name is the name of the lobby.",
                    "type": "string"
                },
                "owner_name": {
                    "description": "OwnerName is the name of the lobby owner.",
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "Open Call to Power Server",
	Description:      "This is the open-source Call to Power and Call to Power 2 server project. This project is not sponsored, maintained or affiliated with Activision.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
