package sofia

import "errors"

const (
	RetOK                 = 100
	RetUnknown            = 101
	RetVersionUnsupported = 102
	RetIllegalRequest     = 103
	RetUserLoggedIn       = 104
	RetUserNotLoggedIn    = 105
)

// Return codes and mappings
var returnCodes = map[int]string{
	RetOK:                 "OK",
	RetUnknown:            "unknown mistake",
	RetVersionUnsupported: "Version not supported",
	RetIllegalRequest:     "Illegal request",
	RetUserLoggedIn:       "The user has logged in",
	RetUserNotLoggedIn:    "The user is not logged in",
	106:                   "username or password is wrong",
	107:                   "No permission",
	108:                   "time out",
	109:                   "Failed to find, no corresponding file found",
	110:                   "Find successful, return all files",
	111:                   "Find success, return some files",
	112:                   "This user already exists",
	113:                   "this user does not exist",
	114:                   "This user group already exists",
	115:                   "This user group does not exist",
	116:                   "Error 116",
	117:                   "Wrong message format",
	118:                   "PTZ protocol not set",
	119:                   "No query to file",
	120:                   "Configure to enable",
	121:                   "MEDIA_CHN_NOT CONNECT digital channel is not connected",
	150:                   "Successful, the device needs to be restarted",
	202:                   "User not logged in",
	203:                   "The password is incorrect",
	204:                   "User illegal",
	205:                   "User is locked",
	206:                   "User is on the blacklist",
	207:                   "Username is already logged in",
	208:                   "Input is illegal",
	209:                   "The index is repeated if the user to be added already exists, etc.",
	210:                   "No object exists, used when querying",
	211:                   "Object does not exist",
	212:                   "Account is in use",
	213:                   "The subset is out of scope (such as the group's permissions exceed the permission table, the user permissions exceed the group's permission range, etc.)",
	214:                   "The password is illegal",
	215:                   "Passwords do not match",
	216:                   "Retain account",
	502:                   "The command is illegal",
	503:                   "Intercom has been turned on",
	504:                   "Intercom is not turned on",
	511:                   "Already started upgrading",
	512:                   "Not starting upgrade",
	513:                   "Upgrade data error",
	514:                   "upgrade unsuccessful",
	515:                   "update successed",
	521:                   "Restore default failed",
	522:                   "Need to restart the device",
	523:                   "Illegal default configuration",
	602:                   "Need to restart the app",
	603:                   "Need to restart the system",
	604:                   "Error writing a file",
	605:                   "Feature not supported",
	606:                   "verification failed",
	607:                   "Configuration does not exist",
	608:                   "Configuration parsing error",
}

// Return string for the return code
func CodeToString(code int) (string, error) {
	txt, found := returnCodes[code]
	if found {
		return txt, nil
	}
	return "", errors.New("unrecognized return code")
}
