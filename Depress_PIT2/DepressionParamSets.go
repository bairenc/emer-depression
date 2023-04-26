package main

import "github.com/emer/emergent/params"

var ParamSets = params.Sets{
	{Name: "Base", Desc: "these are the best params", Sheets: params.Sheets{
	"NetSize": &params.Sheet{
			{Sel: "#Hidden2", Desc: "second hidden layer",
				Params: params.Params{
					"Layer.X": "20",
					"Layer.Y": "5",
				}},
		},
		"Network": &params.Sheet{
			{Sel: "Prjn", Desc: "norm and momentum on works better, but wt bal is not better for smaller nets",
				Params: params.Params{
					"Prjn.Learn.Norm.On":     "true",
					"Prjn.Learn.Momentum.On": "true",
					"Prjn.Learn.WtBal.On":    "false",
					"Prjn.Learn.Learn":       "true",
					"Prjn.Learn.Lrate":       "0.04",
					"Prjn.WtInit.Dist":       "Uniform",
					"Prjn.WtInit.Mean":       "0.5",
					"Prjn.WtInit.Var":        "0.25",
					"Prjn.WtScale.Abs":       "1",
				}},
			{Sel: "#MBAppToApproach", Desc: "Set weight to fixed value of 1",
				Params: params.Params{
					"Prjn.Learn.Learn": "false",
					"Prjn.Learn.Lrate": "0",
					"Prjn.WtInit.Dist": "Uniform",
					"Prjn.WtInit.Mean": "1",
					"Prjn.WtInit.Var":  "0",
					"Prjn.WtScale.Abs": "1",
				}},
			{Sel: "#MBAvToAvoidance", Desc: "Set weight to fixed value of 1",
				Params: params.Params{
					"Prjn.Learn.Learn": "false",
					"Prjn.Learn.Lrate": "0",
					"Prjn.WtInit.Dist": "Uniform",
					"Prjn.WtInit.Mean": "1",
					"Prjn.WtInit.Var":  "0",
					"Prjn.WtScale.Abs": "1", 
				}},
			{Sel: "#VTAToApproach", Desc: "Set weight to fixed value of .5",
				Params: params.Params{
					"Prjn.Learn.Learn": "false",
					"Prjn.Learn.Lrate": "0",
					"Prjn.WtInit.Dist": "Uniform",
					"Prjn.WtInit.Mean": ".5",
					"Prjn.WtInit.Var":  "0",
					"Prjn.WtScale.Abs": "1", // not sure if need a weight scale of 2 here
				}},
			{Sel: "#VTAToAvoidance", Desc: "Set weight to fixed value of .5, inhibitory connection",
				Params: params.Params{
					"Prjn.Learn.Learn": "false",
					"Prjn.Learn.Lrate": "0",
					"Prjn.WtInit.Dist": "Uniform",
					"Prjn.WtInit.Mean": ".5",
					"Prjn.WtInit.Var":  "0",
					"Prjn.WtScale.Abs": "1",
				}},
			{Sel: "#DyDAToApproach", Desc: "Set weight to fixed value of 1, and weight scale of 0.3, inhibitory connection",
				Params: params.Params{
					"Prjn.Learn.Learn": "false",
					"Prjn.Learn.Lrate": "0",
					"Prjn.WtInit.Dist": "Uniform",
					"Prjn.WtInit.Mean": "1",
					"Prjn.WtInit.Var":  "0",
					"Prjn.WtScale.Abs": "0.3", // TEST: dynorphin should not be the main influence, so lower weight scale
				}},
			{Sel: "#DyDAToVTA", Desc: "Set weight to fixed value of 1, and weight scale of 0.5, inhibitory connection",
				Params: params.Params{
					"Prjn.Learn.Learn": "false",
					"Prjn.Learn.Lrate": "0",
					"Prjn.WtInit.Dist": "Uniform",
					"Prjn.WtInit.Mean": "1",
					"Prjn.WtInit.Var":  "0",
					"Prjn.WtScale.Abs": "0.5", // slightly larger effect on VTA than on NAc neuron
				}},
			{Sel: "#CostToBehavior", Desc: "Set weight to fixed value of 1, inhibitory connection",
				Params: params.Params{
					"Prjn.Learn.Learn": "false",
					"Prjn.Learn.Lrate": "0",
					"Prjn.WtInit.Dist": "Uniform",
					"Prjn.WtInit.Mean": ".5", // reducing size of fixed weight to reduce impact of Cost
					"Prjn.WtInit.Var":  "0",
					"Prjn.WtScale.Abs": ".5", // reducing the impact of cost on behavior, since DataTables too high.
				}},
			{Sel: "#Hidden2ToBehavior", Desc: "Set weight to fixed value of 1, inhibitory connection",
				Params: params.Params{
					"Prjn.WtScale.Abs": "1.5", // increasing the impact of Hidden2 on behavior, Behavior hard to activate.
				}},
				
			{Sel: ".Back", Desc: "top-down back-projections MUST have lower relative weight scale, otherwise network hallucinates",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.3",
				}},	
				
			{Sel: "Layer", Desc: "using 2.3 inhib for all of network -- can explore",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "2.3",
					"Layer.Act.Gbar.L":     "0.1",
					"Layer.Act.XX1.Gain":   "100",
				}},
			{Sel: "#Behavior", Desc: "Make Behavior layer selective for 1 behavior",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "2.7",
					"Layer.Act.XX1.Gain":   "400",
					"Layer.Act.Gbar.E":     "1.1",
				}},
			{Sel: "#Approach", Desc: "",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "1.5",
					"Layer.Act.XX1.Gain":   "400", // Gain on activation function
					"Layer.Act.Gbar.E":     "1.0", // conductance
					"Layer.Act.Noise.Dist": "Gaussian",
					"Layer.Act.Noise.Mean": "0.0",
					"Layer.Act.Noise.Var":  "0.0",
					"Layer.Act.Noise.Type": "GeNoise",
				}},
			{Sel: "#Avoidance", Desc: "",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "1.3",
					"Layer.Inhib.Layer.FB": "1.0",
					"Layer.Act.XX1.Thr":    ".49",  // threshold for activation function
					"Layer.Act.XX1.Gain":   "400",  // Gain on activation function
					"Layer.Act.Gbar.E":     "1.35", // conductance higher because Avoidance more sensitive
					"Layer.Act.Noise.Dist": "Gaussian",
					"Layer.Act.Noise.Mean": "0.0",
					"Layer.Act.Noise.Var":  "0.0",
					"Layer.Act.Noise.Type": "GeNoise",
				}},
			{Sel: "#Hidden1", Desc: "Make Hidden representation a bit sparser",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "2.0",
				}},

			{Sel: "#Hidden2", Desc: "Make Hidden representation a bit sparser",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "2.0",
					"Layer.Act.Gbar.E":     "1.0",
				}},
			{Sel: "#VTA", Desc: "Manipulate baseline activation of VTA layer",
				Params: params.Params{
					"Layer.Inhib.Layer.On": "true", // Turn inhibition on
					"Layer.Inhib.Layer.Gi": "1.0",
					"Layer.Act.Init.Decay": "1", 
					"Layer.Act.Init.Ge": "0", // Baseline level of excitatory conductance (Ge), added in as a constant background level of excitatory input
					"Layer.Act.Init.Act": "0", // initial activation value.  This and above don't seem to have any effect during a trial.
					"Layer.Act.Noise.Dist": "Gaussian",
					"Layer.Act.Noise.Mean": "0.4",
					"Layer.Act.Noise.Var":  "0.0",
					"Layer.Act.Noise.Type": "GeNoise",
					"Layer.Act.Noise.Fixed": "true",
					// Make VTA a Hidden layer.
				}},
			{Sel: "#DyDA", Desc: "Manipulate baseline activation of Dynorphin layer",
				Params: params.Params{
					"Layer.Act.Init.Ge": "0", // we need a similar param for the dynorphin layer?
					// for this model DyDA will probably be an input layer
				}},

			
		},
		//		"Sim": &params.Sheet{ // sim params apply to sim object
		//			{Sel: "Sim", Desc: "best params always finish in this time",
		//				Params: params.Params{
		//					"Sim.MaxEpcs": "100",
		//				}},
		//		},
	}},
	// 	},
	//		"Sim": &params.Sheet{ // sim params apply to sim object
	//			{Sel: "Sim", Desc: "takes longer -- generally doesn't finish..",
	//				Params: params.Params{
	//					"Sim.MaxEpcs": "100",
	//				}},
	//		},
	// }},
	// 	{Name: "NoMomentum", Desc: "no momentum or normalization", Sheets: params.Sheets{
	// 		"Network": &params.Sheet{
	// 			{Sel: "Prjn", Desc: "no norm or momentum",
	// 				Params: params.Params{
	// 					"Prjn.Learn.Norm.On":     "false",
	// 					"Prjn.Learn.Momentum.On": "false",
	// 				}},
	// 		},
	// 	}},
	// 	{Name: "WtBalOn", Desc: "try with weight bal on", Sheets: params.Sheets{
	// 		"Network": &params.Sheet{
	// 			{Sel: "Prjn", Desc: "weight bal on",
	// 				Params: params.Params{
	// 					"Prjn.Learn.WtBal.On": "true",
	// 				}},
	// 		},
	// 	}},
	// }
}
