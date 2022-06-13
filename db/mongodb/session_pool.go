/*
 * session_pool.go implements session pool, cache session
 * between MetaNode and DataNode, when someone wants to connect and create
 * session with a specific DataNode(Host:Port), calls pool.GetSession()
 * after finish rpc call, calls pool.ReturnSession() to return this
 * session to session pool.
 * This file implements two kinds of pool: FixedSizeSessionPool and UnfixedSizeSessionPool
 * The difference is: when current session number exceeds quota we set, how they behave
 * FixedSizeSessionPool: wait until someone return a session
 * UnfixedSizeSessionPool: no need to wait, just create a new session and open it
 */

package mongodb

import (
	"gopkg.in/mgo.v2"
	"sync"
	"time"
)

// SessionPool 连接池
type SessionPool interface {
	// get session from pool
	GetSession(hostPort string) (Session, error)
	// return session to pool
	ReturnSession(s Session, status SessionStatus)
}

type Session struct {
	*mgo.Session
	hostPort string
	index    int
	status   internalSessionStatus
}

type internalSessionStatus int

const (
	sessionUnConnect  internalSessionStatus = 0
	sessionBorrowed   internalSessionStatus = 1
	sessionUnBorrowed internalSessionStatus = 2
)

func (s internalSessionStatus) String() string {
	switch s {
	case sessionUnConnect:
		return "UnConnect"
	case sessionBorrowed:
		return "Borrowed"
	case sessionUnBorrowed:
		return "UnBorrowed"
	default:
		return "Unkown"
	}
}

// SessionStatus OK / Error
type SessionStatus int

const (
	SessionError SessionStatus = 0 // for return session
	SessionOK    SessionStatus = 1 // for return session
)

// FixedSizeSessionPool maxConnPerHost最大连接数，如果申请时没有可用连接会等待
// 直到有可用连接后返回
type FixedSizeSessionPool struct {
	sync.Mutex
	sessionMap     map[string]*SessionQueue
	maxConnPerHost int
	timeout        time.Duration
}

func NewFixedSizeSessionPool(maxConnPerHost int, timeout time.Duration) *FixedSizeSessionPool {
	return &FixedSizeSessionPool{
		sessionMap:     make(map[string]*SessionQueue),
		maxConnPerHost: maxConnPerHost,
		timeout:        timeout,
	}
}

func (fp *FixedSizeSessionPool) GetSession(hostPort string) (Session, error) {
	fp.Lock()
	sq, ok := fp.sessionMap[hostPort]

	if !ok {
		sq = NewSessionQueue(fp.maxConnPerHost, fp.timeout, hostPort)
		fp.sessionMap[hostPort] = sq
	}
	fp.Unlock()
	return sq.GetSession()
}

func (fp *FixedSizeSessionPool) ReturnSession(s Session, status SessionStatus) {
	fp.Lock()
	sq, ok := fp.sessionMap[s.hostPort]

	if !ok || sq == nil {
		// glog.Warningf("session pool should exist hostPort=%s", s.hostPort)
		s.Close()
		fp.Unlock()
		return
	}
	fp.Unlock()
	sq.ReturnSession(s, status)
}

type SessionQueue struct {
	m        sync.Mutex
	cond     sync.Cond
	sessions []Session
	maxConn  int
	timeout  time.Duration
	hostPort string
}

func NewSessionQueue(maxConn int, timeout time.Duration, hostPort string) *SessionQueue {
	q := SessionQueue{sessions: make([]Session, maxConn),
		maxConn:  maxConn,
		timeout:  timeout,
		hostPort: hostPort,
	}

	q.cond.L = &q.m
	for i := range q.sessions {
		q.sessions[i].index = i
		q.sessions[i].status = sessionUnConnect
	}

	return &q
}

func (q *SessionQueue) GetSession() (Session, error) {
	q.m.Lock()
	defer q.m.Unlock()
	var i int
	for {
		i = q.getFreeSockIndex()
		if i != -1 {
			break
		}
		q.cond.Wait()
		// glog.V(3).Infof("session queue get session loop")
	}

	if q.sessions[i].status == sessionUnBorrowed {
		q.sessions[i].status = sessionBorrowed
		return q.sessions[i], nil
	} else if q.sessions[i].status == sessionUnConnect {
		s, err := newSession(q.hostPort, q.timeout)
		if err != nil {
			// glog.Warningf("%s", err)
			return s, err
		}
		q.sessions[i] = s
		q.sessions[i].status = sessionBorrowed
		q.sessions[i].index = i
		return q.sessions[i], nil
	}
	panic("sessionQueue should not go here")
}

func (q *SessionQueue) ReturnSession(s Session, status SessionStatus) {
	q.m.Lock()
	defer q.m.Unlock()

	// util.Assert(0 <= s.index && s.index < q.maxConn, "s.index should < q.maxConn")
	// util.Assert(q.sessions[s.index].status == sessionBorrowed, "s.sessions[i].status==borrowed")

	if status == SessionError {
		s.Close()
		q.sessions[s.index].status = sessionUnConnect
		q.sessions[s.index].Session = nil
	} else {
		q.sessions[s.index].status = sessionUnBorrowed
	}

	q.cond.Broadcast()
}

func (q *SessionQueue) getFreeSockIndex() int {
	for i := range q.sessions {
		if q.sessions[i].status == sessionUnBorrowed {
			return i
		}
	}

	for i := range q.sessions {
		if q.sessions[i].status == sessionUnConnect {
			return i
		}
	}
	return -1
}

// SessionPool maxConnPerHost如果申请时没有可用连接, 会创建新的连接
// 归还给连接池时发现连接池满，直接关闭连接
type UnfixedSizeSessionPool struct {
	sync.Mutex
	sessionMap     map[string]chan Session
	maxConnPerHost int
	timeout        time.Duration
}

func NewUnfixedSessionPool(maxConnPerHost int, timeout time.Duration) *UnfixedSizeSessionPool {
	return &UnfixedSizeSessionPool{sessionMap: make(map[string]chan Session),
		maxConnPerHost: maxConnPerHost,
		timeout:        timeout,
	}
}

func (p *UnfixedSizeSessionPool) GetSession(hostPort string) (Session, error) {
	p.Lock()
	sessionChan, ok := p.sessionMap[hostPort]
	p.Unlock()

	if ok {
		return p.getSessionFromChan(hostPort, sessionChan)
	}

	p.Lock()
	sessionChan, ok = p.sessionMap[hostPort]
	if !ok {
		p.sessionMap[hostPort] = make(chan Session, p.maxConnPerHost)
	}
	p.Unlock()

	return p.getSessionFromChan(hostPort, sessionChan)
}

func (p *UnfixedSizeSessionPool) ReturnSession(s Session, status SessionStatus) {
	if status == SessionError {
		s.Close()
		return // notice if session error, close and return
	}

	p.Lock()
	sessionChan, ok := p.sessionMap[s.hostPort]
	p.Unlock()

	if !ok {
		// glog.Warningf("session pool should exist hostPort=%s", s.hostPort)
		s.Close()
	}

	select {
	case sessionChan <- s:
	default:
		s.Close()
	}
}

func newSession(servers string, timeout time.Duration) (Session, error) {
	session, err := mgo.Dial(servers)
	if err != nil {
		return Session{}, err
	}

	session.SetMode(mgo.Monotonic, true)
	// return session, nil
	return Session{Session: session, hostPort: servers, status: sessionUnConnect}, nil
}

func (p *UnfixedSizeSessionPool) getSessionFromChan(hostPort string, sessionChan chan Session) (Session, error) {
	select {
	case s := <-sessionChan:
		return s, nil
	default:
		return newSession(hostPort, p.timeout)
	}
}
