package consensuser

import (
	cont "github.com/s-mx/replob/containers"
	"testing"
	"math/rand"
)

func TestOneNode(t *testing.T) {
	conf := NewMasterlessConfiguration(1)
	carries := cont.NewCarries(1)
	LocalDispatchers := NewLocalDispatchers(1, conf, t)
	dsp := LocalDispatchers[0]

	helper := newTestCommitHelper(1, carries, LocalDispatchers)
	cm := NewTestLocalCommitter(0, helper)
	cons := NewCalmConsensuser(dsp, cm, conf, 0)

	cons.Propose(carries[0])
	if helper.CheckSafety() == false {
		t.Error("Carry isn't committed")
	}
}

func TestTwoNodes(t *testing.T) {
	conf := NewMasterlessConfiguration(2)
	carries := cont.NewCarries(1, 2)
	LocalDispatchers := NewLocalDispatchers(2, conf, t)
	dsp1 := LocalDispatchers[0]
	dsp2 := LocalDispatchers[1]

	helper := newTestCommitHelper(2, carries, LocalDispatchers)
	cm1 := NewTestLocalCommitter(0, helper)
	cm2 := NewTestLocalCommitter(1, helper)
	cons1 := NewCalmConsensuser(dsp1, cm1, conf, 0)
	LocalDispatchers[0].cons = cons1
	cons2 := NewCalmConsensuser(dsp2, cm2, conf, 1)
	LocalDispatchers[1].cons = cons2

	cons1.Propose(carries[0])
	LocalDispatchers[0].proceedFirstMessage(1)
	LocalDispatchers[1].proceedFirstMessage(0)
	LocalDispatchers[0].ClearQueues()
	LocalDispatchers[1].ClearQueues()

	cons2.Propose(carries[1])
	LocalDispatchers[1].proceedFirstMessage(0)
	LocalDispatchers[0].proceedFirstMessage(1)

	if helper.CheckSafety() == false {
		t.Error("Safety is broken")
	}
}


func TestThreeNodes(t *testing.T) {
	conf := NewMasterlessConfiguration(3)
	carries := cont.NewCarries(1, 2)
	LocalDispatchers := NewLocalDispatchers(3, conf, t)
	dsp1 := LocalDispatchers[0]
	dsp2 := LocalDispatchers[1]
	dsp3 := LocalDispatchers[2]

	helper := newTestCommitHelper(3, carries, LocalDispatchers)
	cm1 := NewTestLocalCommitter(0, helper)
	cm2 := NewTestLocalCommitter(1, helper)
	cm3 := NewTestLocalCommitter(2, helper)

	cons1 := NewCalmConsensuser(dsp1, Committer(cm1), conf, 0)
	cons2 := NewCalmConsensuser(dsp2, Committer(cm2), conf, 1)
	cons3 := NewCalmConsensuser(dsp3, Committer(cm3), conf, 2)
	LocalDispatchers[0].cons = cons1
	LocalDispatchers[1].cons = cons2
	LocalDispatchers[2].cons = cons3

	cons1.Propose(carries[0])
	LocalDispatchers[0].proceedFirstMessage(1)
	LocalDispatchers[0].proceedFirstMessage(2)
	LocalDispatchers[1].proceedFirstMessage(0)
	LocalDispatchers[1].proceedFirstMessage(2)
	LocalDispatchers[2].proceedFirstMessage(1)
	LocalDispatchers[2].proceedFirstMessage(0)
	LocalDispatchers[2].proceedFirstMessage(0)

	LocalDispatchers[0].ClearQueues()
	LocalDispatchers[1].ClearQueues()
	LocalDispatchers[2].ClearQueues()

	cons2.Propose(carries[1])
	LocalDispatchers[1].proceedFirstMessage(0)
	LocalDispatchers[1].proceedFirstMessage(2)
	LocalDispatchers[0].proceedFirstMessage(1)
	LocalDispatchers[0].proceedFirstMessage(2)
	LocalDispatchers[2].proceedFirstMessage(1)
	LocalDispatchers[2].proceedFirstMessage(0)

	if helper.CheckSafety() == false {
		t.Error("Carry isn't committed")
	}
}

func RunRandomTest(numberNodes int, numberCarries int, t *testing.T) {
	Source := rand.NewSource(42)
	generator := rand.New(Source)

	conf := NewMasterlessConfiguration(uint32(numberNodes))
	carries := cont.NewCarriesN(numberCarries)
	LocalDispatchers := NewLocalDispatchers(numberNodes, conf, t)

	helper := newTestCommitHelper(numberNodes, carries, LocalDispatchers)
	consensusers := []*CalmConsensuser{}
	for ind := 0; ind < numberNodes; ind++ {
		cm := NewTestLocalCommitter(cont.NodeId(ind), helper)
		cons := NewCalmConsensuser(LocalDispatchers[ind], Committer(cm), conf, cont.NodeId(ind))
		LocalDispatchers[ind].cons = cons
		consensusers = append(consensusers, cons)
	}

	consensusers[0].Propose(carries[0])

	numberProposedCarries := 1
	for numberProposedCarries != numberCarries {
		for true {
			flag := false
			for ind := 0; ind < numberNodes; ind++ {
				if LocalDispatchers[ind].proceedRandomMessage(generator) == true {
					flag = true
				}
			}

			ind := helper.findIndLastCommit(numberProposedCarries)
			if ind != -1 && numberProposedCarries < numberCarries {
				consensusers[ind].Propose(carries[numberProposedCarries])
				numberProposedCarries += 1
				continue
			}

			if flag == false {
				break
			}
		}
	}

	for true {
		flag := false
		for ind := 0; ind < numberNodes; ind++ {
			if LocalDispatchers[ind].proceedRandomMessage(generator) == true {
				flag = true
			}
		}

		if flag == false {
			break
		}
	}

	if helper.CheckSafety() == false {
		t.Error("Carry isn't committed")
	}
}

func TestRandomMessages2(t *testing.T) {
	RunRandomTest(2, 1, t)
}

func TestRandomMessages5(t *testing.T) {
	RunRandomTest(5, 10, t)
}

func TestRandomMessages5_100(t *testing.T) {
	RunRandomTest(5, 100, t)
}

func TestRandomMessages10_10(t *testing.T) {
	RunRandomTest(10, 10, t)
}

func TestRandomMessages10_100(t *testing.T) {
	RunRandomTest(10, 100, t)
}

func TestDisconnectThreeNodes(t *testing.T) {
	conf := NewMasterlessConfiguration(3)
	carries := cont.NewCarries(1)
	LocalDispatchers := NewLocalDispatchers(3, conf, t)
	dsp1 := LocalDispatchers[0]
	dsp2 := LocalDispatchers[1]
	dsp3 := LocalDispatchers[2]

	helper := newTestCommitHelper(3, carries, LocalDispatchers)
	cm1 := NewTestLocalCommitter(0, helper)
	cm2 := NewTestLocalCommitter(1, helper)
	cm3 := NewTestLocalCommitter(2, helper)

	cons1 := NewCalmConsensuser(dsp1, Committer(cm1), conf, 0)
	cons2 := NewCalmConsensuser(dsp2, Committer(cm2), conf, 1)
	cons3 := NewCalmConsensuser(dsp3, Committer(cm3), conf, 2)
	LocalDispatchers[0].cons = cons1
	LocalDispatchers[1].cons = cons2
	LocalDispatchers[2].cons = cons3

	cons1.Propose(carries[0])
	LocalDispatchers[0].proceedFirstMessage(1)
	LocalDispatchers[0].proceedFirstMessage(2)
	LocalDispatchers[0].Stop()
	cons2.OnDisconnect(0)
	cons3.OnDisconnect(0)

	LocalDispatchers[1].proceedFirstMessage(2)
	LocalDispatchers[2].proceedFirstMessage(1)
	LocalDispatchers[1].proceedFirstMessage(2)
	LocalDispatchers[2].proceedFirstMessage(1)

	if helper.CheckSafety() == false {
		t.Error("Carry isn't committed")
	}
}

func RunRandomDisconnectTest(numberNodes int, numberCarries int, numberDisconnects int, t *testing.T) {
	Source := rand.NewSource(42)
	generator := rand.New(Source)

	conf := NewMasterlessConfiguration(uint32(numberNodes))
	carries := cont.NewCarriesN(numberCarries)
	LocalDispatchers := NewLocalDispatchers(numberNodes, conf, t)

	subsetDisconnectedNodes := cont.NewRandomSubset(conf.Info, numberDisconnects, generator)
	helper := newTestCommitHelper(numberNodes, carries, LocalDispatchers)
	consensusers := []*CalmConsensuser{}
	for ind := 0; ind < numberNodes; ind++ {
		cm := NewTestLocalCommitter(cont.NodeId(ind), helper)
		cons := NewCalmConsensuser(LocalDispatchers[ind], Committer(cm), conf, cont.NodeId(ind))
		LocalDispatchers[ind].cons = cons
		consensusers = append(consensusers, cons)
	}

	consensusers[0].Propose(carries[0])

	numberProposedCarries := 1
	for numberProposedCarries != numberCarries {
		for true {
			flag := false
			// FIXME: consider checking majority separately
			// Disconnect this
			if subsetDisconnectedNodes.Size() > 0 &&  generator.Float32() < 0.01 {
				indDisconnect := subsetDisconnectedNodes.Get(0)
				subsetDisconnectedNodes.Erase(indDisconnect)


				LocalDispatchers[indDisconnect].Stop()
				for ind := 0; ind < numberNodes; ind++ {
					if LocalDispatchers[ind].IsRunning() {
						consensusers[ind].OnDisconnect(cont.NodeId(indDisconnect))
					}
				}
			}

			for ind := 0; ind < numberNodes; ind++ {
				if LocalDispatchers[ind].proceedRandomMessage(generator) == true {
					flag = true
				}
			}

			ind := helper.findIndLastCommit(numberProposedCarries)
			if ind != -1 && numberProposedCarries < numberCarries {
				consensusers[ind].Propose(carries[numberProposedCarries])
				numberProposedCarries += 1
				continue
			}

			if flag == false {
				break
			}
		}
	}

	for true {
		flag := false
		for ind := 0; ind < numberNodes; ind++ {
			if LocalDispatchers[ind].proceedRandomMessage(generator) == true {
				flag = true
			}
		}

		if flag == false {
			break
		}
	}

	if helper.CheckSafety() == false {
		t.Error("Carry isn't committed")
	}
}

func TestRandomDisconnect5(t *testing.T) {
	RunRandomDisconnectTest(5, 10, 2, t)
}


func TestRandomDisconnect5_100(t *testing.T) {
	RunRandomDisconnectTest(5, 100, 2, t)
}

func TestRandomDisconnect10_10(t *testing.T) {
	RunRandomDisconnectTest(10, 10, 4, t)
}

func TestRandomDisconnect10_100(t *testing.T) {
	RunRandomDisconnectTest(10, 100, 4, t)
}

/*
Tests TODO:
1. Disconnect + liveness checks.
	+ change safety check: all prefixes with the same length must be the same
	+ no disconnects && no message drops => all lengths must be the same
	+ minor disconnnects without message drops => there are majority nodes with desired messages
	-! on drop message: just check for prefix safety
	- on limit dropped message on each step: full safety check
2. Propose must be right after commit.
 */
