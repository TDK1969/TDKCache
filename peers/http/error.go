package http_server

type Err struct {
	Error     string `json:"error"`
	ErrorCode string `json:"error_code"`
}

type ErrorResponse struct {
	HttpSC int
	Error  Err
}

var (
	ErrorRequestBodyParseFailed = ErrorResponse{
		HttpSC: 400,
		Error: Err{
			Error:     "Request body is not correct.",
			ErrorCode: "001",
		},
	}
	ErrorNotAuthUser = ErrorResponse{
		HttpSC: 401,
		Error: Err{
			Error:     "User authentication failed.",
			ErrorCode: "002",
		},
	}
	ErrorDuplicatedLoginName = ErrorResponse{
		HttpSC: 500,
		Error: Err{
			Error:     "Username duplicated.",
			ErrorCode: "003",
		},
	}
	ErrorDBError = ErrorResponse{
		HttpSC: 500,
		Error: Err{
			Error:     "DB ops failed.",
			ErrorCode: "004",
		},
	}
	ErrorInternalFaults = ErrorResponse{
		HttpSC: 500,
		Error: Err{
			Error:     "Internal service error.",
			ErrorCode: "005",
		},
	}
	ErrorURLParamsParseFailed = ErrorResponse{
		HttpSC: 400,
		Error: Err{
			Error:     "Request URL params are not correct.",
			ErrorCode: "006",
		},
	}
	ErrorTooManyRequests = ErrorResponse{
		HttpSC: 429,
		Error: Err{
			Error:     "Too many requests",
			ErrorCode: "007",
		},
	}
	ErrorUploadFileTooBig = ErrorResponse{
		HttpSC: 400,
		Error: Err{
			Error:     "File too big",
			ErrorCode: "008",
		},
	}
	ErrorGroupUnexists = ErrorResponse{
		HttpSC: 404,
		Error: Err{
			Error:     "No such group",
			ErrorCode: "008",
		},
	}
)
