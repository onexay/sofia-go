package sofia

/*
 *
 */
type Session struct {
	id    byte   // Session ID
	idStr string // Session ID as string
	pkts  uint32 // Packets handled
}

/*
 *
 */
func NewSesion(id byte, idStr string) *Session {
	// Allocate a new session
	session := new(Session)

	session.id = id
	session.idStr = idStr

	return session
}

/*
 *
 */
func DeleteSession(s *Session) {
}

func (s *Session) ID() byte {
	return s.id
}

func (s *Session) IDStr() *string {
	return &s.idStr
}

/*
 *
 */
func (s *Session) SystemInfo() {

}
