package sofia

/*
 *
 */
type Session struct {
	id   byte // Session index (provided by Server)
	pkts uint // Packets exchanged
}

/*
 *
 */
func NewSesion(id byte) *Session {
	session := new(Session)

	session.id = id
	session.pkts = 0

	return session
}

/*
 *
 */
func DeleteSession() {

}

/*
 *
 */
func (session *Session) SysInfoRes(data []byte) {

}
