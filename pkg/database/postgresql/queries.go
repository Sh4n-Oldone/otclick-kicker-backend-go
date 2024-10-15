package postgresql

const (
	queryTemplate string = `
		SELECT * 
		FROM api_operations.template
		(
		    pemail := $1,
    		ppassword := $2,
			pnickname := $3,
			prole_id := $4
		);`
)
