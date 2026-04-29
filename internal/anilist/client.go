package anilist

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type graphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type GraphQLResponse struct {
	Data struct {
		Media struct {
			ID    int `json:"id"`
			Title struct {
				English string `json:"english"`
				Romaji  string `json:"romaji"`
				Native  string `json:"native"`
			} `json:"title"`
		} `json:"Media"`
	} `json:"data"`
}

func getAnilistData(title string) (GraphQLResponse, error) {
	reqBody := graphQLRequest{
		Query: `
        query ($search: String) {
            Media(search: $search, type: ANIME) {
                id
                title {
                    english
                    romaji
                    native
                }
            }
        }
    `,
		Variables: map[string]interface{}{
			"search": title,
		},
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return GraphQLResponse{}, err
	}

	req, err := http.NewRequest(
		"POST",
		"https://graphql.anilist.co",
		bytes.NewReader(jsonBytes),
	)
	if err != nil {
		return GraphQLResponse{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return GraphQLResponse{}, err
	}
	defer resp.Body.Close()

	var response GraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return GraphQLResponse{}, err
	}

	return response, nil
}

func getAnilistDataById(id int) (GraphQLResponse, error) {
	reqBody := graphQLRequest{
		Query: `
        query ($id: Int) {
            Media(id: $id, type: ANIME) {
                id
                title {
                    english
                    romaji
                    native
                }
            }
        }
    `,
		Variables: map[string]interface{}{
			"id": id,
		},
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return GraphQLResponse{}, err
	}

	req, err := http.NewRequest(
		"POST",
		"https://graphql.anilist.co",
		bytes.NewReader(jsonBytes),
	)
	if err != nil {
		return GraphQLResponse{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return GraphQLResponse{}, err
	}
	defer resp.Body.Close()

	var response GraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return GraphQLResponse{}, err
	}

	return response, nil
}
