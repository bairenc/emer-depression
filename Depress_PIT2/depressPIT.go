// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// this is the framework for the Depression model, based on the ra25 example
// ra25 runs a simple random-associator four-layer leabra network
// that uses the standard supervised learning paradigm to learn
// mappings between 25 random input / output patterns
// defined over 5x5 input / output layers (i.e., 25 units)
package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/emer/emergent/egui"
	"github.com/emer/emergent/elog"
	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/env"
	"github.com/emer/emergent/estats"
	"github.com/emer/emergent/etime"
	"github.com/emer/emergent/netview"
	"github.com/emer/emergent/relpos"
	_ "github.com/emer/emergent/patgen"
	"github.com/emer/emergent/prjn"
	"github.com/emer/etable/agg"
	"github.com/emer/etable/etable"
	_ "github.com/emer/etable/etensor"
	_ "github.com/emer/etable/etview" // include to get gui views
	"github.com/emer/etable/split"
	"github.com/emer/leabra/leabra"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/gimain"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
	"github.com/goki/mat32"
)

func main() {
	TheSim.New()
	TheSim.Config()
	if len(os.Args) > 1 {
		TheSim.CmdArgs() // simple assumption is that any args = no gui -- could add explicit arg if you want
	} else {
		gimain.Main(func() { // this starts gui -- requires valid OpenGL display connection (e.g., X11)
			guirun()
		})
	}
}

func guirun() {
	TheSim.Init()
	win := TheSim.ConfigGui()
	win.StartEventLoop()
}

// LogPrec is precision for saving float values in logs
const LogPrec = 4

// Sim encapsulates the entire simulation model, and we define all the
// functionality as methods on this struct.  This structure keeps all relevant
// state information organized and available without having to pass everything around
// as arguments to methods, and provides the core GUI interface (note the view tags
// for the fields which provide hints to how things should be displayed).
type Sim struct {
	Net          *leabra.Network  `view:"no-inline" desc:"the network -- click to view / edit parameters for layers, prjns, etc"`
	Params       emer.Params      `view:"inline" desc:"all parameter management"`
	Instr        *etable.Table     `view:"no-inline" desc:"Training pattern for Instrumental Learning"`
	Pvlv         *etable.Table     `view:"no-inline" desc:"Training pattern for Pavlovian Learning"`
	Trn    		 *etable.Table     `view:"no-inline" desc:"Table that controls type of training and number of Epochs of training"`
	TestData 	 *etable.Table     `view:"no-inline" desc:"Table for the Test data file"`
	Training	 string			   `view:"no-inline" desc:"Type of training: Pavlovian or Instrumental"`
	Tag          string           `desc:"extra tag string to add to any file names output from sim (e.g., weights files, log files, params for run)"`
	Stats        estats.Stats     `desc:"contains computed statistic values"`
	LayNms		 []string		  `desc:"Names of Layers to which Training and Testing values are applied"`
	Logs         elog.Logs        `desc:"Contains all the logs and information about the logs.'"`
	StartRun     int              `desc:"starting run number -- typically 0 but can be set in command args for parallel runs on a cluster"`
	MaxRuns      int              `desc:"maximum number of model runs to perform (starting from StartRun)"`
	MaxEpcs      int              `desc:"maximum number of epochs to run per model run"`
	NZeroStop    int              `desc:"if a positive number, training will stop after this many epochs with zero SSE"`
	TrainEnv     env.FixedTable   `desc:"Training environment -- contains everything about iterating over input / output patterns over training"`
	TestEnv      env.FixedTable   `desc:"Testing environment -- manages iterating over testing"`
	Time         leabra.Time      `desc:"leabra timing parameters and state"`
	ViewUpdt     netview.ViewUpdt `view:"inline" desc:"netview update parameters"`
	TestInterval int              `desc:"how often to run through all the test patterns, in terms of training epochs -- can use 0 or -1 for no testing"`
	PCAInterval  int              `desc:"how frequently (in epochs) to compute PCA on hidden representations to measure variance?"`

	GUI          egui.GUI         `view:"-" desc:"manages all the gui elements"`
	SaveWts      bool             `view:"-" desc:"for command-line run only, auto-save final weights after each run"`
	NoGui        bool             `view:"-" desc:"if true, runing in no GUI mode"`
	LogSetParams bool             `view:"-" desc:"if true, print message for all params that are set"`
	NeedsNewRun  bool             `view:"-" desc:"flag to initialize NewRun if last one finished"`
	RndSeeds     []int64          `view:"-" desc:"a list of random seeds to use for each run"`
	NetData      *netview.NetData `view:"-" desc:"net data for recording in nogui mode"`
}

// this registers this Sim Type and gives it properties that e.g.,
// prompt for filename for save methods.
var KiT_Sim = kit.Types.AddType(&Sim{}, SimProps)

// TheSim is the overall state for this simulation
var TheSim Sim

// New creates new blank elements and initializes defaults
func (ss *Sim) New() {
	ss.Net = &leabra.Network{}
	ss.Instr = &etable.Table{}
	ss.Pvlv = &etable.Table{}
	ss.Trn = &etable.Table{}
	ss.TestData = &etable.Table{}
	ss.Params.Params = ParamSets
	ss.Params.AddNetwork(ss.Net)
	ss.Params.AddSim(ss)
	ss.Params.AddNetSize()
	ss.Stats.Init()
	ss.RndSeeds = make([]int64, 100) // make enough for plenty of runs
	for i := 0; i < 100; i++ {
		ss.RndSeeds[i] = int64(i) + 1 // exclude 0
	}
	ss.TestInterval = -1
	ss.PCAInterval = 5
	ss.Time.Defaults()
}

////////////////////////////////////////////////////////////////////////////////////////////
// 		Configs

// Config configures all the elements using the standard functions
func (ss *Sim) Config() {
	//ss.ConfigPats()
	ss.OpenPats()
	ss.ConfigEnv()
	ss.ConfigNet(ss.Net)
	ss.ConfigLogs()
}

func (ss *Sim) ConfigEnv() {
	if ss.MaxRuns == 0 { // allow user override
		ss.MaxRuns = 1
	}
	if ss.MaxEpcs == 0 { // allow user override
		ss.MaxEpcs = 100
		ss.NZeroStop = -1
	}

	ss.TrainEnv.Nm = "TrainEnv"
	ss.TrainEnv.Dsc = "training params and state"
	ss.TrainEnv.Table = etable.NewIdxView(ss.Instr)
	ss.TrainEnv.Validate()
	ss.TrainEnv.Run.Max = ss.MaxRuns // note: we are not setting epoch max -- do that manually

	ss.TestEnv.Nm = "TestEnv"
	ss.TestEnv.Dsc = "testing params and state"
	ss.TestEnv.Table = etable.NewIdxView(ss.TestData)
	ss.TestEnv.Sequential = true
	ss.TestEnv.Validate()

	// note: to create a train / test split of pats, do this:
	// all := etable.NewIdxView(ss.Pats)
	// splits, _ := split.Permuted(all, []float64{.8, .2}, []string{"Train", "Test"})
	// ss.TrainEnv.Table = splits.Splits[0]
	// ss.TestEnv.Table = splits.Splits[1]

	ss.TrainEnv.Init(0)
	ss.TestEnv.Init(0)
}

func (ss *Sim) ConfigNet(net *leabra.Network) {
	ss.LayNms = []string{"EnviroFeatures", "InteroState","MBApp","MBAv","Approach","Avoidance","Behavior","Cost","DyDA"}
	ss.Params.AddLayers([]string{"Hidden1", "Hidden2"}, "Hidden")
	ss.Params.SetObject("NetSize")

	net.InitName(net, "Depress")

	envfeats := net.AddLayer2D("EnviroFeatures", 1, 8, emer.Input)
	interostate := net.AddLayer2D("InteroState", 1, 8, emer.Input)
	motivebias1 := net.AddLayer2D("MBApp", 1, 5, emer.Input)
	motivebias2 := net.AddLayer2D("MBAv", 1, 3, emer.Input)
	approach := net.AddLayer2D("Approach", 1, 5, emer.Target)
	avoidance := net.AddLayer2D("Avoidance", 1, 3, emer.Target)
	out := net.AddLayer2D("Behavior", 1, 16, emer.Target)
	
	vta := net.AddLayer2D("VTA", 1, 1, emer.Hidden)  // Make VTA a hidden layer with baseline self activatioin
	cost := net.AddLayer2D("Cost", 1, 16, emer.Input)
	
	//Capturing stress-dynorphin expression and down-regulation of DA
	dyn := net.AddLayer2D("DyDA", 1, 1, emer.Input)

	hid1 := net.AddLayer2D("Hidden1", ss.Params.LayY("Hidden1", 5), ss.Params.LayX("Hidden1", 20), emer.Hidden)
	hid2 := net.AddLayer2D("Hidden2", ss.Params.LayY("Hidden2", 5), ss.Params.LayX("Hidden2", 20), emer.Hidden)

	

	// use this to position layers relative to each other
	// default is Above, YAlign = Front, XAlign = Center
	// hid2.SetRelPos(relpos.Rel{Rel: relpos.RightOf, Other: "Hidden1", YAlign: relpos.Front, Space: 2})

	// note: see emergent/prjn module for all the options on how to connect
	// NewFull returns a new prjn.Full connectivity pattern
	full := prjn.NewFull()
    one2one := prjn.NewOneToOne()
    
	net.ConnectLayers(envfeats, hid1, full, emer.Forward)
	net.ConnectLayers(interostate, hid1, full, emer.Forward)
	net.ConnectLayers(motivebias1, approach, one2one, emer.Forward)
	net.ConnectLayers(motivebias2, avoidance, one2one, emer.Forward)
	
	net.ConnectLayers(vta, approach, full, emer.Forward)

	//feedfoward connections between layers

	net.BidirConnectLayers(hid1, approach, full)
	net.BidirConnectLayers(hid1, avoidance, full)
	net.BidirConnectLayers(approach, hid2, full)
	net.BidirConnectLayers(avoidance, hid2, full)
	net.BidirConnectLayers(hid2, out, full)
	
	

	//bidirectional connections between layers

	net.ConnectLayers(vta, avoidance, full, emer.Inhib)
	
	// one to one since cost is specific to each behavior
	net.ConnectLayers(cost, out, one2one, emer.Inhib)
	
	
	net.ConnectLayers(dyn, approach, full, emer.Inhib)
	net.ConnectLayers(dyn, vta, full, emer.Inhib)

	//inhibitory connections between layers



// Commands that are used to position the layers in the Netview
	
	hid1.SetRelPos(relpos.Rel{Rel: relpos.Above, Other: "EnviroFeatures", YAlign: relpos.Front, XAlign: relpos.Right, XOffset: 1})
	approach.SetRelPos(relpos.Rel{Rel: relpos.Above, Other: "Hidden1", YAlign: relpos.Front, XAlign: relpos.Right, XOffset: 1})
	hid2.SetRelPos(relpos.Rel{Rel: relpos.Above, Other: "Approach", YAlign: relpos.Front, XAlign: relpos.Left, YOffset: 0})
	out.SetRelPos(relpos.Rel{Rel: relpos.Above, Other: "Hidden2", YAlign: relpos.Front, XAlign: relpos.Right, XOffset: 1})
	
	interostate.SetRelPos(relpos.Rel{Rel: relpos.RightOf, Other: "EnviroFeatures", YAlign: relpos.Front, XAlign: relpos.Right, XOffset: 1})
	avoidance.SetRelPos(relpos.Rel{Rel: relpos.RightOf, Other: "Approach", YAlign: relpos.Front, XAlign: relpos.Right, XOffset: 1})
	motivebias1.SetRelPos(relpos.Rel{Rel: relpos.RightOf, Other: "Avoidance", YAlign: relpos.Front, XAlign: relpos.Right, XOffset: 1})
	motivebias2.SetRelPos(relpos.Rel{Rel: relpos.RightOf, Other: "MBApp", YAlign: relpos.Front, XAlign: relpos.Right, XOffset: 1})
	vta.SetRelPos(relpos.Rel{Rel: relpos.RightOf, Other: "Hidden2", YAlign: relpos.Front, XAlign: relpos.Right, XOffset: 1})
	cost.SetRelPos(relpos.Rel{Rel: relpos.RightOf, Other: "Behavior", YAlign: relpos.Front, XAlign: relpos.Right, XOffset: 1})
	dyn.SetRelPos(relpos.Rel{Rel: relpos.LeftOf, Other: "Approach", YAlign: relpos.Front, XAlign: relpos.Right})

	// note: can set these to do parallel threaded computation across multiple cpus
	// not worth it for this small of a model, but definitely helps for larger ones
	// if Thread {
	// 	hid2.SetThread(1)
	// 	out.SetThread(1)
	// }

	// note: if you wanted to change a layer type from e.g., Target to Compare, do this:
	// out.SetType(emer.Compare)
	// that would mean that the output layer doesn't reflect target values in plus phase
	// and thus removes error-driven learning -- but stats are still computed.

	net.Defaults()
	ss.Params.SetObject("Network")
	err := net.Build()
	if err != nil {
		log.Println(err)
		return
	}
	net.InitWts()
}


func (ss *Sim) OpenPats() {
	ss.Instr.OpenCSV("DepressInstr2.tsv", etable.Tab) // Instrumental training data
	ss.Pvlv.OpenCSV("DepressPvlv.tsv", etable.Tab) // Pavlovian training data
	ss.Trn.OpenCSV("InstrThenPvlv.tsv", etable.Tab) // Order of training and number of epochs for each, 
	// Pavlov first or Instrumental first. Should eventually create menu to choose.
	ss.TestData.OpenCSV("DepressTest.tsv", etable.Tab) // Test data
}

////////////////////////////////////////////////////////////////////////////////
// 	    Init, utils

// Init restarts the run, and initializes everything, including network weights
// and resets the epoch log table
func (ss *Sim) Init() {
	ss.InitRndSeed()
	ss.ConfigEnv() // re-config env just in case a different set of patterns was
	// selected or patterns have been modified etc
	ss.GUI.StopNow = false
	ss.Params.SetMsg = ss.LogSetParams
	ss.Params.SetAll()
	ss.NewRun()
	ss.ViewUpdt.Update()
}

// InitRndSeed initializes the random seed based on current training run number
func (ss *Sim) InitRndSeed() {
	run := ss.TrainEnv.Run.Cur
	rand.Seed(ss.RndSeeds[run])
}

// NewRndSeed gets a new set of random seeds based on current time -- otherwise uses
// the same random seeds for every run
func (ss *Sim) NewRndSeed() {
	rs := time.Now().UnixNano()
	for i := 0; i < 100; i++ {
		ss.RndSeeds[i] = rs + int64(i)
	}
}

////////////////////////////////////////////////////////////////////////////////
// 	    Running the Network, starting bottom-up..

// AlphaCyc runs one alpha-cycle (100 msec, 4 quarters) of processing.
// External inputs must have already been applied prior to calling,
// using ApplyExt method on relevant layers (see TrainTrial, TestTrial).
// If train is true, then learning DWt or WtFmDWt calls are made.
// Handles netview updating within scope of AlphaCycle
func (ss *Sim) AlphaCyc(train bool) {
	// ss.Win.PollEvents() // this can be used instead of running in a separate goroutine
	ss.Net.AlphaCycInit(train)
	ss.Time.AlphaCycStart()
	for qtr := 0; qtr < 4; qtr++ {
		for cyc := 0; cyc < ss.Time.CycPerQtr; cyc++ {
			ss.Net.Cycle(&ss.Time)
			ss.StatCounters(train)
			if !train {
				ss.Log(etime.Test, etime.Cycle)
			}
			ss.Time.CycleInc()
			ss.ViewUpdt.UpdateCycle(cyc)
		}
		ss.Net.QuarterFinal(&ss.Time)
		ss.Time.QuarterInc()
		ss.ViewUpdt.UpdateTime(etime.GammaCycle)
	}
	ss.StatCounters(train)

	if train {
		ss.Net.DWt()
		ss.ViewUpdt.RecordSyns() // note: critical to update weights here so DWt is visible
		ss.Net.WtFmDWt()
	}
	ss.ViewUpdt.UpdateTime(etime.AlphaCycle)
	if !train {
		ss.GUI.UpdatePlot(etime.Test, etime.Cycle) // make sure always updated at end
	}
}

// ApplyInputs applies input patterns from given environment.
// It is good practice to have this be a separate method with appropriate
// args so that it can be used for various different contexts
// (training, testing, etc).
func (ss *Sim) ApplyInputs(en env.Env) {
	// ss.Net.InitExt() // clear any existing inputs -- not strictly necessary if always
	// going to the same layers, but good practice and cheap anyway

	lays := ss.LayNms
	for _, lnm := range lays {
		ly := ss.Net.LayerByName(lnm).(leabra.LeabraLayer).AsLeabra()
		pats := en.State(ly.Nm)
		if pats != nil {
			ly.ApplyExt(pats)
		}
	}
}

// TrainTrial runs one trial of training using TrainEnv
func (ss *Sim) TrainTrial() {
	if ss.NeedsNewRun {
		ss.NewRun()
	}

	ss.TrainEnv.Step() // the Env encapsulates and manages all counter state

	// Key to query counters FIRST because current state is in NEXT epoch
	// if epoch counter has changed
	epc, _, chg := ss.TrainEnv.Counter(env.Epoch)
	if chg {
		if (ss.PCAInterval > 0) && ((epc-1)%ss.PCAInterval == 0) { // -1 so runs on first epc
			ss.PCAStats()
		}
		ss.Log(etime.Train, etime.Epoch)
		ss.ViewUpdt.UpdateTime(etime.Epoch)
		if ss.TestInterval > 0 && epc%ss.TestInterval == 0 { // note: epc is *next* so won't trigger first time
			ss.TestAll()
		}
		if epc >= ss.MaxEpcs || (ss.NZeroStop > 0 && ss.Stats.Int("NZero") >= ss.NZeroStop) {
			// done with training..
			ss.RunEnd()
			if ss.TrainEnv.Run.Incr() { // we are done!
				ss.GUI.StopNow = true
				return
			} else {
				ss.NeedsNewRun = true
				return
			}
		}
	}

	ss.ApplyInputs(&ss.TrainEnv)
	ss.AlphaCyc(true) // train
	ss.TrialStats()
	ss.Log(etime.Train, etime.Trial)
	if (ss.PCAInterval > 0) && (epc%ss.PCAInterval == 0) {
		ss.Log(etime.Analyze, etime.Trial)
	}
}

// RunEnd is called at the end of a run -- save weights, record final log, etc here
func (ss *Sim) RunEnd() {
	ss.Log(etime.Train, etime.Run)
	if ss.SaveWts {
		fnm := ss.WeightsFileName()
		fmt.Printf("Saving Weights to: %s\n", fnm)
		ss.Net.SaveWtsJSON(gi.FileName(fnm))
	}
}

// NewRun intializes a new run of the model, using the TrainEnv.Run counter
// for the new run value
func (ss *Sim) NewRun() {
	ss.InitRndSeed()
	run := ss.TrainEnv.Run.Cur
	ss.TrainEnv.Init(run)
	ss.TestEnv.Init(run)
	ss.Time.Reset()
	ss.Net.SaveWtsJSON("trained.wts")
	ss.Net.InitWts()
	ss.InitStats()
	ss.StatCounters(true)
	ss.Logs.ResetLog(etime.Train, etime.Epoch)
	ss.Logs.ResetLog(etime.Test, etime.Epoch)
	ss.NeedsNewRun = false
}

// TrainEpoch runs training trials for remainder of this epoch
func (ss *Sim) TrainEpoch() {
	ss.GUI.StopNow = false
	curEpc := ss.TrainEnv.Epoch.Cur
	for {
		ss.TrainTrial()
		if ss.GUI.StopNow || ss.TrainEnv.Epoch.Cur != curEpc {
			break
		}
	}
	ss.Stopped()
}

// TrainRun runs training trials for remainder of run
func (ss *Sim) TrainRun() {
	ss.GUI.StopNow = false
	curRun := ss.TrainEnv.Run.Cur
	for {
		ss.TrainTrial()
		if ss.GUI.StopNow || ss.TrainEnv.Run.Cur != curRun {
			break
		}
	}
	ss.Stopped()
}

// Train runs the full training from this point onward
func (ss *Sim) Train() {
	ss.GUI.StopNow = false
	for {
		ss.TrainTrial()
		if ss.GUI.StopNow {
			break
		}
	}
	ss.Stopped()
}

// Stop tells the sim to stop running
func (ss *Sim) Stop() {
	ss.GUI.StopNow = true
}

// Stopped is called when a run method stops running -- updates the IsRunning flag and toolbar
func (ss *Sim) Stopped() {
	ss.GUI.Stopped()
}

// SaveWeights saves the network weights -- when called with giv.CallMethod
// it will auto-prompt for filename
func (ss *Sim) SaveWeights(filename gi.FileName) {
	ss.Net.SaveWtsJSON(filename)
}

func (ss *Sim) TrainPIT() {

for i := 0; i < 2; i++ {
		
ss.Training = ss.Trn.CellString("Training", i)
ss.MaxEpcs = int(ss.Trn.CellFloat("MaxEpoch", i))



switch ss.Training {

	case "INSTRUMENTAL":

	ss.TrainEnv.Epoch.Cur = 0  //set current epoch to 0 so that training starts from 0 epochs
	
	// Need to set up training patterns.  	
	ss.TrainEnv.Table = etable.NewIdxView(ss.Instr)
	ss.TestEnv.Table = etable.NewIdxView(ss.TestData)
	
	run := ss.TrainEnv.Run.Cur  //  Is Init what is resetting the TrainEnv table index properly?
	ss.TrainEnv.Init(run)

// Unlesion Hidden and Behavior layer to make sure all layers are unlesioned
	ss.Net.LayerByName("Hidden2").SetOff(false)
	ss.Net.LayerByName("Behavior").SetOff(false)

// Load saved weights
// OpenWtsJSON opens trained weights
		ss.Net.OpenWtsJSON("trained.wts")
	

// Define Environment and InteroState as Compare layers
// Define Approach and Avoid as Input layers
// Define Behavior as a Target layer
	ss.Net.LayerByName("EnviroFeatures").SetType(emer.Input)
	ss.Net.LayerByName("InteroState").SetType(emer.Input)
	ss.Net.LayerByName("Approach").SetType(emer.Input)
	ss.Net.LayerByName("Avoidance").SetType(emer.Input)
	ss.Net.LayerByName("Behavior").SetType(emer.Target)
	

// Lesion Environment, InteroState, and Hidden2 layers
	ss.Net.LayerByName("EnviroFeatures").SetOff(true)
	ss.Net.LayerByName("InteroState").SetOff(true)
	ss.Net.LayerByName("Hidden1").SetOff(true)

	ss.Train()


// Unlesion Environment and InteroState layers
	ss.Net.LayerByName("EnviroFeatures").SetOff(false)
	ss.Net.LayerByName("InteroState").SetOff(false)
	ss.Net.LayerByName("Hidden1").SetOff(false)

// Save weights
	ss.Net.SaveWtsJSON("trained.wts")



// Define Environment and InteroState as Input layers
// Define Approach and Avoid as Target layers
// Define Behavior as a Target layer
// Makes sure that layers are set to default.
	ss.Net.LayerByName("EnviroFeatures").SetType(emer.Input)
	ss.Net.LayerByName("InteroState").SetType(emer.Input)
	ss.Net.LayerByName("Approach").SetType(emer.Target)
	ss.Net.LayerByName("Avoidance").SetType(emer.Target)	
	ss.Net.LayerByName("Behavior").SetType(emer.Target)

	case "PAVLOV":

	ss.TrainEnv.Epoch.Cur = 0  //set current epoch to 0 so that training starts from 0 epochs
	
// Pavlovian Training

// Need to set up training patterns  

	ss.TrainEnv.Table = etable.NewIdxView(ss.Pvlv)
	ss.TestEnv.Table = etable.NewIdxView(ss.TestData)	
	
	run := ss.TrainEnv.Run.Cur  //  Is Init what is resetting the TrainEnv table index properly?
	ss.TrainEnv.Init(run)

	



// Define Environment and InteroState as Input layers
// Define Approach and Avoid as Target layers
// Define Behavior as a Compare layer
	ss.Net.LayerByName("EnviroFeatures").SetType(emer.Input)
	ss.Net.LayerByName("InteroState").SetType(emer.Input)
	ss.Net.LayerByName("Approach").SetType(emer.Target)
	ss.Net.LayerByName("Avoidance").SetType(emer.Target)
	ss.Net.LayerByName("Behavior").SetType(emer.Target)

// Load saved weights
		ss.Net.OpenWtsJSON("trained.wts")
	

// Lesion Hidden layer and Behavior layer
	ss.Net.LayerByName("Hidden2").SetOff(true)
	ss.Net.LayerByName("Behavior").SetOff(true)


// Train until number of Epochs of training reached
	ss.Train()

//Unlesion Hidden Layer and Behavior layer
	ss.Net.LayerByName("Hidden2").SetOff(false)
	ss.Net.LayerByName("Behavior").SetOff(false)


// Save weights
	ss.Net.SaveWtsJSON("trained.wts")
	
	
// Define Environment and InteroState as Input layers
// Define Approach and Avoid as Target layers
// Define Behavior as a Target layer
// Makes sure that layers are set to default.

	ss.Net.LayerByName("EnviroFeatures").SetType(emer.Input)
	ss.Net.LayerByName("InteroState").SetType(emer.Input)
	ss.Net.LayerByName("Approach").SetType(emer.Target)
	ss.Net.LayerByName("Avoidance").SetType(emer.Target)	
	ss.Net.LayerByName("Behavior").SetType(emer.Target)

	
	
		}
	}
}









////////////////////////////////////////////////////////////////////////////////////////////
// Testing

// TestTrial runs one trial of testing -- always sequentially presented inputs
func (ss *Sim) TestTrial(returnOnChg bool) {
	ss.TestEnv.Step()

	// Query counters FIRST
	_, _, chg := ss.TestEnv.Counter(env.Epoch)
	if chg {
		ss.ViewUpdt.UpdateTime(etime.Epoch)
		ss.Log(etime.Test, etime.Epoch)
		if returnOnChg {
			return
		}
	}

	ss.ApplyInputs(&ss.TestEnv)
	ss.AlphaCyc(false) // !train
	ss.TrialStats()
	ss.Log(etime.Test, etime.Trial)
	if ss.NetData != nil { // offline record net data from testing, just final state
		ss.NetData.Record(ss.ViewUpdt.Text, -1, 1)
	}
}

// TestItem tests given item which is at given index in test item list
func (ss *Sim) TestItem(idx int) {
	cur := ss.TestEnv.Trial.Cur
	ss.TestEnv.Trial.Cur = idx
	ss.TestEnv.SetTrialName()
	ss.ApplyInputs(&ss.TestEnv)
	ss.AlphaCyc(false) // !train
	ss.TrialStats()
	ss.TestEnv.Trial.Cur = cur
}

// TestAll runs through the full set of testing items
func (ss *Sim) TestAll() {
	ss.TestEnv.Init(ss.TrainEnv.Run.Cur)
	for {
		ss.TestTrial(true) // return on change -- don't wrap
		_, _, chg := ss.TestEnv.Counter(env.Epoch)
		if chg || ss.GUI.StopNow {
			break
		}
	}
}

// RunTestAll runs through the full set of testing items, has stop running = false at end -- for gui
func (ss *Sim) RunTestAll() {
	ss.GUI.StopNow = false
	ss.TestAll()
	ss.Stopped()
}
/*
func (ss *Sim) ConfigPats() {
	dt := ss.Pats
	dt.SetMetaData("name", "TrainPats")
	dt.SetMetaData("desc", "Training patterns")
	sch := etable.Schema{
		{"Name", etensor.STRING, nil, nil},
		{"Input", etensor.FLOAT32, []int{5, 5}, []string{"Y", "X"}},
		{"Output", etensor.FLOAT32, []int{5, 5}, []string{"Y", "X"}},
	}
	dt.SetFromSchema(sch, 25)

	patgen.PermutedBinaryRows(dt.Cols[1], 6, 1, 0)
	patgen.PermutedBinaryRows(dt.Cols[2], 6, 1, 0)
	dt.SaveCSV("random_5x5_25_gen.tsv", etable.Tab, etable.Headers)
}
*/

////////////////////////////////////////////////////////////////////////////////////////////
// 		Logging

// RunName returns a name for this run that combines Tag and Params -- add this to
// any file names that are saved.
func (ss *Sim) RunName() string {
	rn := ""
	if ss.Tag != "" {
		rn += ss.Tag + "_"
	}
	rn += ss.Params.Name()
	if ss.StartRun > 0 {
		rn += fmt.Sprintf("_%03d", ss.StartRun)
	}
	return rn
}

// RunEpochName returns a string with the run and epoch numbers with leading zeros, suitable
// for using in weights file names.  Uses 3, 5 digits for each.
func (ss *Sim) RunEpochName(run, epc int) string {
	return fmt.Sprintf("%03d_%05d", run, epc)
}

// WeightsFileName returns default current weights file name
func (ss *Sim) WeightsFileName() string {
	return ss.Net.Nm + "_" + ss.RunName() + "_" + ss.RunEpochName(ss.TrainEnv.Run.Cur, ss.TrainEnv.Epoch.Cur) + ".wts"
}

// LogFileName returns default log file name
func (ss *Sim) LogFileName(lognm string) string {
	return ss.Net.Nm + "_" + ss.RunName() + "_" + lognm + ".tsv"
}

// InitStats initializes all the statistics.
// called at start of new run
func (ss *Sim) InitStats() {
	// clear rest just to make Sim look initialized
	ss.Stats.SetFloat("TrlErr", 0.0)
	ss.Stats.SetFloat("TrlSSE", 0.0)
	ss.Stats.SetFloat("TrlAvgSSE", 0.0)
	ss.Stats.SetFloat("TrlCosDiff", 0.0)
	ss.Stats.SetInt("FirstZero", -1) // critical to reset to -1
	ss.Stats.SetInt("NZero", 0)
}

// StatCounters saves current counters to Stats, so they are available for logging etc
// Also saves a string rep of them to the GUI, if the GUI is active
func (ss *Sim) StatCounters(train bool) {
	ev := ss.TrainEnv
	if !train {
		ev = ss.TestEnv
	}
	ss.Stats.SetInt("Run", ss.TrainEnv.Run.Cur)
	ss.Stats.SetInt("Epoch", ss.TrainEnv.Epoch.Cur)
	ss.Stats.SetInt("Trial", ev.Trial.Cur)
	ss.Stats.SetString("TrialName", ev.TrialName.Cur)
	ss.Stats.SetInt("Cycle", ss.Time.Cycle)
	ss.ViewUpdt.Text = ss.Stats.Print([]string{"Run", "Epoch", "Trial", "TrialName", "Cycle", "AvgSSE", "TrlErr", "TrlCosDiff"})
}

// TrialStats computes the trial-level statistics.
// Aggregation is done directly from log data.
func (ss *Sim) TrialStats() {
	out := ss.Net.LayerByName("Behavior").(leabra.LeabraLayer).AsLeabra()

	sse, avgsse := out.MSE(0.5) // 0.5 = per-unit tolerance -- right side of .5
	ss.Stats.SetFloat("TrlSSE", sse)
	ss.Stats.SetFloat("TrlAvgSSE", avgsse)
	ss.Stats.SetFloat("TrlCosDiff", float64(out.CosDiff.Cos))

	if sse > 0 {
		ss.Stats.SetFloat("TrlErr", 1)
	} else {
		ss.Stats.SetFloat("TrlErr", 0)
	}
}

//////////////////////////////////////////////
//  Logging

func (ss *Sim) ConfigLogs() {
	ss.ConfigLogItems()
	ss.Logs.CreateTables()
	ss.Logs.SetContext(&ss.Stats, ss.Net)
	// don't plot certain combinations we don't use
	ss.Logs.NoPlot(etime.Train, etime.Cycle)
	ss.Logs.NoPlot(etime.Test, etime.Run)
	// note: Analyze not plotted by default
	ss.Logs.SetMeta(etime.Train, etime.Run, "LegendCol", "Params")
}

// Log is the main logging function, handles special things for different scopes
func (ss *Sim) Log(mode etime.Modes, time etime.Times) {
	dt := ss.Logs.Table(mode, time)
	row := dt.Rows
	switch {
	case mode == etime.Test && time == etime.Epoch:
		ss.LogTestErrors()
	case time == etime.Cycle:
		row = ss.Stats.Int("Cycle")
	case time == etime.Trial:
		row = ss.Stats.Int("Trial")
	}

	ss.Logs.LogRow(mode, time, row) // also logs to file, etc
	if time == etime.Cycle {
		ss.GUI.UpdateCyclePlot(etime.Test, ss.Time.Cycle)
	} else {
		ss.GUI.UpdatePlot(mode, time)
	}

	switch {
	case mode == etime.Train && time == etime.Run:
		ss.LogRunStats()
	}
}

// LogTestErrors records all errors made across TestTrials, at Test Epoch scope
func (ss *Sim) LogTestErrors() {
	sk := etime.Scope(etime.Test, etime.Trial)
	lt := ss.Logs.TableDetailsScope(sk)
	ix, _ := lt.NamedIdxView("TestErrors")
	ix.Filter(func(et *etable.Table, row int) bool {
		return et.CellFloat("Err", row) > 0 // include error trials
	})
	ss.Logs.MiscTables["TestErrors"] = ix.NewTable()

	allsp := split.All(ix)
	split.Agg(allsp, "SSE", agg.AggSum)
	// note: can add other stats to compute
	ss.Logs.MiscTables["TestErrorStats"] = allsp.AggsToTable(etable.AddAggName)
}

// LogRunStats records stats across all runs, at Train Run scope
func (ss *Sim) LogRunStats() {
	sk := etime.Scope(etime.Train, etime.Run)
	lt := ss.Logs.TableDetailsScope(sk)
	ix, _ := lt.NamedIdxView("RunStats")

	spl := split.GroupBy(ix, []string{"Params"})
	split.Desc(spl, "FirstZero")
	split.Desc(spl, "PctCor")
	ss.Logs.MiscTables["RunStats"] = spl.AggsToTable(etable.AddAggName)
}

// PCAStats computes PCA statistics on recorded hidden activation patterns
// from Analyze, Trial log data
func (ss *Sim) PCAStats() {
	ss.Stats.PCAStats(ss.Logs.IdxView(etime.Analyze, etime.Trial), "ActM", ss.Net.LayersByClass("Hidden"))
	ss.Logs.ResetLog(etime.Analyze, etime.Trial)
}

////////////////////////////////////////////////////////////////////////////////////////////
// 		Gui

// ConfigGui configures the GoGi gui interface for this simulation,
func (ss *Sim) ConfigGui() *gi.Window {
	title := "Depression Model"
	ss.GUI.MakeWindow(ss, "Depress", title, `Model of role of dopamine in depression. See <a href="https://github.com/emer/emergent">emergent on GitHub</a>.</p>`)
	ss.GUI.CycleUpdateInterval = 10

	nv := ss.GUI.AddNetView("NetView")
	nv.SetNet(ss.Net)
	ss.ViewUpdt.Config(nv, etime.AlphaCycle, etime.AlphaCycle)
	ss.GUI.ViewUpdt = &ss.ViewUpdt

	nv.Scene().Camera.Pose.Pos.Set(0, 1, 2.75) // more "head on" than default which is more "top down"
	nv.Scene().Camera.LookAt(mat32.Vec3{0, 0, 0}, mat32.Vec3{0, 1, 0})
	ss.GUI.AddPlots(title, &ss.Logs)

	ss.GUI.AddToolbarItem(egui.ToolbarItem{Label: "Init", Icon: "update",
		Tooltip: "Initialize everything including network weights, and start over.  Also applies current params.",
		Active:  egui.ActiveStopped,
		Func: func() {
			ss.Init()
			ss.GUI.UpdateWindow()
		},
	})
	
		
	ss.GUI.AddToolbarItem(egui.ToolbarItem{Label: "TrainPIT",
		Icon:    "run",
		Tooltip: "Starts PIT training.",
		Active:  egui.ActiveStopped,
		Func: func() {
			if !ss.GUI.IsRunning {
				ss.GUI.IsRunning = true
				ss.GUI.ToolBar.UpdateActions()
				go ss.TrainPIT()
			}
		},
	})
		
	ss.GUI.AddToolbarItem(egui.ToolbarItem{Label: "Train",
		Icon:    "run",
		Tooltip: "Starts the network training, picking up from wherever it may have left off.  If not stopped, training will complete the specified number of Runs through the full number of Epochs of training, with testing automatically occuring at the specified interval.",
		Active:  egui.ActiveStopped,
		Func: func() {
			if !ss.GUI.IsRunning {
				ss.GUI.IsRunning = true
				ss.GUI.ToolBar.UpdateActions()
				go ss.Train()
			}
		},
	})
	ss.GUI.AddToolbarItem(egui.ToolbarItem{Label: "Stop",
		Icon:    "stop",
		Tooltip: "Interrupts running.  Hitting Train again will pick back up where it left off.",
		Active:  egui.ActiveRunning,
		Func: func() {
			ss.Stop()
		},
	})
	ss.GUI.AddToolbarItem(egui.ToolbarItem{Label: "Step Trial",
		Icon:    "step-fwd",
		Tooltip: "Advances one training trial at a time.",
		Active:  egui.ActiveStopped,
		Func: func() {
			if !ss.GUI.IsRunning {
				ss.GUI.IsRunning = true
				ss.TrainTrial()
				ss.GUI.IsRunning = false
				ss.GUI.UpdateWindow()
			}
		},
	})
	ss.GUI.AddToolbarItem(egui.ToolbarItem{Label: "Step Epoch",
		Icon:    "fast-fwd",
		Tooltip: "Advances one epoch (complete set of training patterns) at a time.",
		Active:  egui.ActiveStopped,
		Func: func() {
			if !ss.GUI.IsRunning {
				ss.GUI.IsRunning = true
				ss.GUI.ToolBar.UpdateActions()
				go ss.TrainEpoch()
			}
		},
	})
	ss.GUI.AddToolbarItem(egui.ToolbarItem{Label: "Step Run",
		Icon:    "fast-fwd",
		Tooltip: "Advances one full training Run at a time.",
		Active:  egui.ActiveStopped,
		Func: func() {
			if !ss.GUI.IsRunning {
				ss.GUI.IsRunning = true
				ss.GUI.ToolBar.UpdateActions()
				go ss.TrainRun()
			}
		},
	})

	////////////////////////////////////////////////
	ss.GUI.ToolBar.AddSeparator("test")
	ss.GUI.AddToolbarItem(egui.ToolbarItem{Label: "Test Trial",
		Icon:    "fast-fwd",
		Tooltip: "Runs the next testing trial.",
		Active:  egui.ActiveStopped,
		Func: func() {
			if !ss.GUI.IsRunning {
				ss.GUI.IsRunning = true
				ss.TestTrial(false) // don't return on change -- wrap
				ss.GUI.IsRunning = false
				ss.GUI.UpdateWindow()
			}
		},
	})
	ss.GUI.AddToolbarItem(egui.ToolbarItem{Label: "Test Item",
		Icon:    "step-fwd",
		Tooltip: "Prompts for a specific input pattern name to run, and runs it in testing mode.",
		Active:  egui.ActiveStopped,
		Func: func() {

			gi.StringPromptDialog(ss.GUI.ViewPort, "", "Test Item",
				gi.DlgOpts{Title: "Test Item", Prompt: "Enter the Name of a given input pattern to test (case insensitive, contains given string."},
				ss.GUI.Win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
					dlg := send.(*gi.Dialog)
					if sig == int64(gi.DialogAccepted) {
						val := gi.StringPromptDialogValue(dlg)
						idxs := []int{0} //TODO: //ss.TestEnv.Table.RowsByString("Name", val, etable.Contains, etable.IgnoreCase)
						if len(idxs) == 0 {
							gi.PromptDialog(nil, gi.DlgOpts{Title: "Name Not Found", Prompt: "No patterns found containing: " + val}, gi.AddOk, gi.NoCancel, nil, nil)
						} else {
							if !ss.GUI.IsRunning {
								ss.GUI.IsRunning = true
								fmt.Printf("testing index: %d\n", idxs[0])
								ss.TestItem(idxs[0])
								ss.GUI.IsRunning = false
								ss.GUI.ViewPort.SetNeedsFullRender()
							}
						}
					}
				})

		},
	})
	ss.GUI.AddToolbarItem(egui.ToolbarItem{Label: "Test All",
		Icon:    "step-fwd",
		Tooltip: "Prompts for a specific input pattern name to run, and runs it in testing mode.",
		Active:  egui.ActiveStopped,
		Func: func() {
			if !ss.GUI.IsRunning {
				ss.GUI.IsRunning = true
				ss.GUI.ToolBar.UpdateActions()
				go ss.RunTestAll()
			}
		},
	})

	////////////////////////////////////////////////
	ss.GUI.ToolBar.AddSeparator("log")
	ss.GUI.AddToolbarItem(egui.ToolbarItem{Label: "Reset RunLog",
		Icon:    "reset",
		Tooltip: "Reset the accumulated log of all Runs, which are tagged with the ParamSet used",
		Active:  egui.ActiveAlways,
		Func: func() {
			ss.Logs.ResetLog(etime.Train, etime.Run)
			ss.GUI.UpdatePlot(etime.Train, etime.Run)
		},
	})
	////////////////////////////////////////////////
	ss.GUI.ToolBar.AddSeparator("misc")
	ss.GUI.AddToolbarItem(egui.ToolbarItem{Label: "New Seed",
		Icon:    "new",
		Tooltip: "Generate a new initial random seed to get different results.  By default, Init re-establishes the same initial seed every time.",
		Active:  egui.ActiveAlways,
		Func: func() {
			ss.NewRndSeed()
		},
	})
	ss.GUI.AddToolbarItem(egui.ToolbarItem{Label: "README",
		Icon:    "file-markdown",
		Tooltip: "Opens your browser on the README file that contains instructions for how to run this model.",
		Active:  egui.ActiveAlways,
		Func: func() {
			gi.OpenURL("https://github.com/emer/leabra/blob/master/examples/ra25/README.md")
		},
	})
	ss.GUI.FinalizeGUI(false)
	return ss.GUI.Win
}

// These props register Save methods so they can be used
var SimProps = ki.Props{
	"CallMethods": ki.PropSlice{
		{"SaveWeights", ki.Props{
			"desc": "save network weights to file",
			"icon": "file-save",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"ext": ".wts,.wts.gz",
				}},
			},
		}},
	},
}

func (ss *Sim) CmdArgs() {
	ss.NoGui = true
	var nogui bool
	var saveEpcLog bool
	var saveRunLog bool
	var saveNetData bool
	var note string
	flag.StringVar(&ss.Params.ExtraSets, "params", "", "ParamSet name to use -- must be valid name as listed in compiled-in params or loaded params")
	flag.StringVar(&ss.Tag, "tag", "", "extra tag to add to file names saved from this run")
	flag.StringVar(&note, "note", "", "user note -- describe the run params etc")
	flag.IntVar(&ss.StartRun, "run", 0, "starting run number -- determines the random seed -- runs counts from there -- can do all runs in parallel by launching separate jobs with each run, runs = 1")
	flag.IntVar(&ss.MaxRuns, "runs", 10, "number of runs to do (note that MaxEpcs is in paramset)")
	flag.BoolVar(&ss.LogSetParams, "setparams", false, "if true, print a record of each parameter that is set")
	flag.BoolVar(&ss.SaveWts, "wts", false, "if true, save final weights after each run")
	flag.BoolVar(&saveEpcLog, "epclog", true, "if true, save train epoch log to file")
	flag.BoolVar(&saveRunLog, "runlog", true, "if true, save run epoch log to file")
	flag.BoolVar(&saveNetData, "netdata", false, "if true, save network activation etc data from testing trials, for later viewing in netview")
	flag.BoolVar(&nogui, "nogui", true, "if not passing any other args and want to run nogui, use nogui")
	flag.Parse()
	ss.Init()

	if note != "" {
		fmt.Printf("note: %s\n", note)
	}
	if ss.Params.ExtraSets != "" {
		fmt.Printf("Using ParamSet: %s\n", ss.Params.ExtraSets)
	}

	if saveEpcLog {
		fnm := ss.LogFileName("epc")
		ss.Logs.SetLogFile(etime.Train, etime.Epoch, fnm)
	}
	if saveRunLog {
		fnm := ss.LogFileName("run")
		ss.Logs.SetLogFile(etime.Train, etime.Run, fnm)
	}
	if saveNetData {
		ss.NetData = &netview.NetData{}
		ss.NetData.Init(ss.Net, 200, true) // 200 = amount to save
	}
	if ss.SaveWts {
		fmt.Printf("Saving final weights per run\n")
	}
	fmt.Printf("Running %d Runs starting at %d\n", ss.MaxRuns, ss.StartRun)
	ss.TrainEnv.Run.Set(ss.StartRun)
	ss.TrainEnv.Run.Max = ss.StartRun + ss.MaxRuns
	ss.NewRun()
	ss.Train()

	ss.Logs.CloseLogFiles()

	if saveNetData {
		ndfn := ss.Net.Nm + "_" + ss.RunName() + ".netdata.gz"
		ss.NetData.SaveJSON(gi.FileName(ndfn))
	}
}
