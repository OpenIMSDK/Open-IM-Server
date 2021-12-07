package gate

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/token_verify"
	"Open_IM/pkg/utils"
	"bytes"
	"encoding/gob"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type UserConn struct {
	*websocket.Conn
	w *sync.Mutex
}
type WServer struct {
	wsAddr       string
	wsMaxConnNum int
	wsUpGrader   *websocket.Upgrader
	wsConnToUser map[*UserConn]string
	wsUserToConn map[string]*UserConn
}

func (ws *WServer) onInit(wsPort int) {
	ws.wsAddr = ":" + utils.IntToString(wsPort)
	ws.wsMaxConnNum = config.Config.LongConnSvr.WebsocketMaxConnNum
	ws.wsConnToUser = make(map[*UserConn]string)
	ws.wsUserToConn = make(map[string]*UserConn)
	ws.wsUpGrader = &websocket.Upgrader{
		HandshakeTimeout: time.Duration(config.Config.LongConnSvr.WebsocketTimeOut) * time.Second,
		ReadBufferSize:   config.Config.LongConnSvr.WebsocketMaxMsgLen,
		CheckOrigin:      func(r *http.Request) bool { return true },
	}
}

func (ws *WServer) run() {
	http.HandleFunc("/", ws.wsHandler)         //Get request from client to handle by wsHandler
	err := http.ListenAndServe(ws.wsAddr, nil) //Start listening
	if err != nil {
		log.ErrorByKv("Ws listening err", "", "err", err.Error())
	}
}

func (ws *WServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	if ws.headerCheck(w, r) {
		query := r.URL.Query()
		conn, err := ws.wsUpGrader.Upgrade(w, r, nil) //Conn is obtained through the upgraded escalator
		if err != nil {
			log.ErrorByKv("upgrade http conn err", "", "err", err)
			return
		} else {
			//Connection mapping relationship,
			//userID+" "+platformID->conn

			//Initialize a lock for each user
			newConn := &UserConn{conn, new(sync.Mutex)}
			ws.addUserConn(query["sendID"][0], int32(utils.StringToInt64(query["platformID"][0])), newConn, query["token"][0])
			go ws.readMsg(newConn)
		}
	}
}

func (ws *WServer) readMsg(conn *UserConn) {
	for {
		messageType, msg, err := conn.ReadMessage()
		if messageType == websocket.PingMessage {
			log.NewInfo("", "this is a  pingMessage")
		}
		if err != nil {
			log.ErrorByKv("WS ReadMsg error", "", "userIP", conn.RemoteAddr().String(), "userUid", ws.getUserUid(conn), "error", err)
			ws.delUserConn(conn)
			return
		} else {
			//log.ErrorByKv("test", "", "msgType", msgType, "userIP", conn.RemoteAddr().String(), "userUid", ws.getUserUid(conn))
		}
		ws.msgParse(conn, msg)
		//ws.writeMsg(conn, 1, chat)
	}

}
func (ws *WServer) writeMsg(conn *UserConn, a int, msg []byte) error {
	conn.w.Lock()
	defer conn.w.Unlock()
	return conn.WriteMessage(a, msg)

}
func (ws *WServer) MultiTerminalLoginChecker(uid string, platformID int32, newConn *UserConn, token string) {
	switch config.Config.MultiLoginPolicy {
	case constant.AllLoginButSameTermKick:
		if oldConn, ok := ws.wsUserToConn[genMapKey(uid, platformID)]; ok {
			log.NewDebug("", uid, platformID, "kick old conn")
			ws.sendKickMsg(oldConn, newConn)
			m, err := db.DB.GetTokenMapByUidPid(uid, constant.PlatformIDToName(platformID))
			if err != nil {
				log.NewError("", "get token from redis err", err.Error())
				return
			}
			if m == nil {
				log.NewError("", "get token from redis err", "m is nil")
				return
			}
			for k, _ := range m {
				if k != token {
					m[k] = constant.KickedToken
				}
			}
			log.NewDebug("get map is ", m)
			err = db.DB.SetTokenMapByUidPid(uid, platformID, m)
			if err != nil {
				log.NewError("", "SetTokenMapByUidPid err", err.Error())
				return
			}
			err = oldConn.Close()
			delete(ws.wsConnToUser, oldConn)
			if err != nil {
				log.NewError("", "conn close err", err.Error())
			}
		} else {
			log.NewDebug("no other conn", ws.wsUserToConn)
		}

	case constant.SingleTerminalLogin:
	case constant.WebAndOther:
	}
}
func (ws *WServer) sendKickMsg(oldConn, newConn *UserConn) {
	mReply := Resp{
		ReqIdentifier: constant.WSKickOnlineMsg,
		ErrCode:       constant.ErrTokenInvalid.ErrCode,
		ErrMsg:        constant.ErrTokenInvalid.ErrMsg,
	}
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err := enc.Encode(mReply)
	if err != nil {
		log.NewError(mReply.OperationID, mReply.ReqIdentifier, mReply.ErrCode, mReply.ErrMsg, "Encode Msg error", oldConn.RemoteAddr().String(), newConn.RemoteAddr().String(), err.Error())
		return
	}
	err = ws.writeMsg(oldConn, websocket.BinaryMessage, b.Bytes())
	if err != nil {
		log.NewError(mReply.OperationID, mReply.ReqIdentifier, mReply.ErrCode, mReply.ErrMsg, "WS WriteMsg error", oldConn.RemoteAddr().String(), newConn.RemoteAddr().String(), err.Error())
	}
}
func (ws *WServer) addUserConn(uid string, platformID int32, conn *UserConn, token string) {
	key := genMapKey(uid, platformID)
	rwLock.Lock()
	defer rwLock.Unlock()
	ws.MultiTerminalLoginChecker(uid, platformID, conn, token)
	ws.wsConnToUser[conn] = key
	ws.wsUserToConn[key] = conn
	log.WarnByKv("WS Add operation", "", "wsUser added", ws.wsUserToConn, "uid", uid, "online_num", len(ws.wsUserToConn))

}

func (ws *WServer) delUserConn(conn *UserConn) {
	rwLock.Lock()
	defer rwLock.Unlock()
	var uidPlatform string
	if uid, ok := ws.wsConnToUser[conn]; ok {
		uidPlatform = uid
		if _, ok = ws.wsUserToConn[uid]; ok {
			delete(ws.wsUserToConn, uid)
			log.WarnByKv("WS delete operation", "", "wsUser deleted", ws.wsUserToConn, "uid", uid, "online_num", len(ws.wsUserToConn))
		} else {
			log.WarnByKv("uid not exist", "", "wsUser deleted", ws.wsUserToConn, "uid", uid, "online_num", len(ws.wsUserToConn))
		}
		delete(ws.wsConnToUser, conn)
	}
	err := conn.Close()
	if err != nil {
		log.ErrorByKv("close err", "", "uid", uidPlatform)
	}

}

func (ws *WServer) getUserConn(uid string) *UserConn {
	rwLock.RLock()
	defer rwLock.RUnlock()
	if conn, ok := ws.wsUserToConn[uid]; ok {
		return conn
	}
	return nil
}
func (ws *WServer) getUserUid(conn *UserConn) string {
	rwLock.RLock()
	defer rwLock.RUnlock()

	if uid, ok := ws.wsConnToUser[conn]; ok {
		return uid
	}
	return ""
}
func (ws *WServer) headerCheck(w http.ResponseWriter, r *http.Request) bool {
	status := http.StatusUnauthorized
	query := r.URL.Query()
	if len(query["token"]) != 0 && len(query["sendID"]) != 0 && len(query["platformID"]) != 0 {
		if ok, err := token_verify.VerifyToken(query["token"][0], query["sendID"][0]); !ok {
			e := err.(*constant.ErrInfo)
			log.ErrorByKv("Token verify failed", "", "query", query)
			w.Header().Set("Sec-Websocket-Version", "13")
			http.Error(w, e.ErrMsg, int(e.ErrCode))
			return false
		} else {
			log.InfoByKv("Connection Authentication Success", "", "token", query["token"][0], "userID", query["sendID"][0])
			return true
		}
	} else {
		log.ErrorByKv("Args err", "", "query", query)
		w.Header().Set("Sec-Websocket-Version", "13")
		http.Error(w, http.StatusText(status), status)
		return false
	}

}
func genMapKey(uid string, platformID int32) string {
	return uid + " " + constant.PlatformIDToName(platformID)
}
