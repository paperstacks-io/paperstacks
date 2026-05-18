package domain_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/user/domain"
)

func TestAndi(t *testing.T) {
	service := domain.NewUserService(nil, "https://040d94f7-bedd-43d4-b0cf-36c38688acdd.hanko.io", nil)

	token := "eyJhbGciOiJSUzI1NiIsImtpZCI6IjEzODdkZDY5LTYzNDYtNDM0MS1iZWNjLWYzZGMzMTdmMWVmNyIsInR5cCI6IkpXVCJ9.eyJhbXIiOlsicHdkIl0sImF1ZCI6WyJsb2NhbGhvc3QiXSwiZW1haWwiOnsiYWRkcmVzcyI6ImFuZHJlYXMuYmF1ZXJAdGgtbnVlcm5iZXJnLmRlIiwiaXNfcHJpbWFyeSI6dHJ1ZSwiaXNfdmVyaWZpZWQiOnRydWV9LCJleHAiOjE3NzkyNzE1NTMsImlhdCI6MTc3ODY2Njc1MywiaXNzIjoiaHR0cHM6Ly8wNDBkOTRmNy1iZWRkLTQzZDQtYjBjZi0zNmMzODY4OGFjZGQuaGFua28uaW8iLCJzZXNzaW9uX2lkIjoiYTVmOTA2NmYtMDVmNS00M2Y2LTgxYjgtMjc4NWFkZjc5NTU1Iiwic3ViIjoiNjc5ZGEzMGMtYWI0NC00ZmM5LTkxZTktMzdlZWRiMjEwZjljIn0.CNcDbdnjpaDWgjODfSr6UCldmT2CBRUhqOjnIafphEM2SKt1Qnimos6VxV8Yr1O4V38RQRoQNb1QWjxFMqsYoI5BakqEoXBREQvcNY_IvwZzGtYBYU6_lZm_6Sq4DN3yt7Pn_-m-gRJ3_Zoj2tnpEHYn2vBt0v69J8HPmtv0H72uR0s2pqClTLxVlBHWj99uROxtF26FqwwfmOAg7On5n16GtYADPpJerxpHFAJDO202v6vuRKQTPHes-Xv1Sz7-fODI624xI-JxMDLiSJrOJd2jjeDbwx-_yYx0YEeLEwbuiXXH8KZknXrinEh1qiAEPJUeKHogNmDWytZqmqtp68PPH6sog0COPckvBQOzN0YfZ4v-PtMUb-PL0bT_emHiwMn8JTNowCWb08PrTD9UZrylSiGzAsLZhBA3owBW2QqpPRcR1UMLlj0c9LFCYZKiChgQbbr4-565dod04TcU8G2065PCmfEMYICEZlpl1rS_GrvwNCMgrwPZI9wtR4mtLIIFkQwZKQX1EUJSgtAVhOrGPketaOfgb6lFqW8Ydkus5nQGdq-whuNiah0Ax5U4WU1LC8uukHuYQkMIDd-a5rlm5TOjPtNLkCj2LvN0PP3fuJ2tXYooCofJNhNetVvQXP13KSwcebHnA0JdrDT6cDjpd-OFCDEHQCIUsGYellk"
	user, err := service.CreateFromToken(context.Background(), token)
	if err != nil {
		t.Fatalf("error = %v", err)
	}

	fmt.Println("user.externalID:" + user.ExternalID)
	fmt.Println("user.Email:" + user.Email)
	t.Fail()
}

func TestMeResponseToUser(t *testing.T) {
	t.Parallel()

	jsonRes := `
	{
  "id": "c339547d-e17d-4ba7-8a1d-b3d5a4d17c1c",
  "user_id": "c339547d-e17d-4ba7-8a1d-b3d5a4d17c1c",
  "emails": [
    {
      "id": "5333cc5b-c7c4-48cf-8248-9c184ac72b65",
      "address": "john.doe@example.com",
      "is_verified": true,
      "is_primary": false
    }
  ],
  "created_at": "2023-11-07T05:31:56Z",
  "updated_at": "2023-11-07T05:31:56Z",
  "passkeys": [
    {
      "id": "5333cc5b-c7c4-48cf-8248-9c184ac72b65",
      "name": "iCloud",
      "public_key": "pQECYyagASFYIBblARCP_at3cmprjzQN1lJ...",
      "attestation_type": "packed",
      "aaguid": "01020304-0506-0708-0102-030405060708",
      "last_used_at": "2026-02-24T21:40:36.26936Z",
      "created_at": "2026-02-24T21:40:36.26936Z",
      "transports": [
        "internal"
      ],
      "backup_eligible": true,
      "backup_state": true,
      "mfa_only": false
    }
  ],
  "security_keys": [
    {
      "id": "f826013e-e7e3-4366-a6d8-9359effc8cdd",
      "name": "Yubikey Bio",
      "public_key": "aNMEEyadASFYIBblARCP_at3cmp4gg3zQN1lJ...",
      "attestation_type": "packed",
      "aaguid": "90636e1f-ef82-43bf-bdcf-5255f139d12f",
      "last_used_at": "2026-02-24T21:40:36.26936Z",
      "created_at": "2026-02-24T21:40:36.26936Z",
      "transports": [
        "usb"
      ],
      "backup_eligible": true,
      "backup_state": false,
      "mfa_only": true
    }
  ],
  "metadata": {
    "public_metadata": {
      "role": "admin"
    },
    "unsafe_metadata": {
      "birthday": "2025-05-12"
    }
  },
  "name": "<string>",
  "given_name": "<string>",
  "family_name": "<string>",
  "picture": "<string>",
  "mfa_config": {
    "auth_app_set_up": true,
    "totp_enabled": true,
    "security_key_enabled": true
  },
  "username": {
    "id": "c339547d-e17d-4ba7-8a1d-b3d5a4d17c1c",
    "created_at": "2023-11-07T05:31:56Z",
    "updated_at": "2023-11-07T05:31:56Z",
    "username": "john_doe"
  }
}
	`

	var res domain.MeResponse
	json.Unmarshal([]byte(jsonRes), &res)
	user := res.ToUser()

	if user.Email != "john.doe@example.com" {
		t.Fatalf("user email = %q, want %q", user.Email, "john.doe@example.com")
	}

	if user.ExternalID != "c339547d-e17d-4ba7-8a1d-b3d5a4d17c1c" {
		t.Fatalf("user externalID = %q, want %q", user.ExternalID, "c339547d-e17d-4ba7-8a1d-b3d5a4d17c1c")
	}
}
