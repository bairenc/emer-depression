
package main

import "github.com/emer/emergent/params"

var ParamSets = params.Sets{
	{Name: "Base", Desc: "these are the best params", Sheets: params.Sheets{
		"Network": &params.Sheet{
			{Sel: "Prjn", Desc: "norm and momentum on works better, but wt bal is not better for smaller nets",
				Params: params.Params{
					"Prjn.Learn.Norm.On":     "true",
					"Prjn.Learn.Momentum.On": "true",
					"Prjn.Learn.WtBal.On":    "false",
					"Prjn.Learn.Learn": "true",
					"Prjn.Learn.Lrate":  "0.04",
					"Prjn.WtInit.Dist": "Uniform",
					"Prjn.WtInit.Mean": "0.5",
					"Prjn.WtInit.Var":  "0.25",
					"Prjn.WtScale.Abs": "1",
					
				}},
				{Sel: "#MotiveBiasToApproach", Desc: "Set weight to fixed value of 1",
				Params: params.Params{
					"Prjn.Learn.Learn": "false",
					"Prjn.Learn.Lrate":  "0",
					"Prjn.WtInit.Dist": "Uniform",
					"Prjn.WtInit.Mean": "1",
					"Prjn.WtInit.Var":  "0",
					"Prjn.WtScale.Abs": "1",		
				}},
				{Sel: "#MotiveBiasToAvoid", Desc: "Set weight to fixed value of 1",
				Params: params.Params{
					"Prjn.Learn.Learn": "false",
					"Prjn.Learn.Lrate":  "0",
					"Prjn.WtInit.Dist": "Uniform",
					"Prjn.WtInit.Mean": "1",
					"Prjn.WtInit.Var":  "0",
					"Prjn.WtScale.Abs": "1",		
				}},
				{Sel: "#VTA_DAToApproach", Desc: "Set weight to fixed value of 1, and weight scale of 2",
				Params: params.Params{
					"Prjn.Learn.Learn": "false",
					"Prjn.Learn.Lrate":  "0",
					"Prjn.WtInit.Dist": "Uniform",
					"Prjn.WtInit.Mean": "1",
					"Prjn.WtInit.Var":  "0",
					"Prjn.WtScale.Abs": "2",		
				}},
				{Sel: "#VTA_DAToAvoid", Desc: "Set weight to fixed value of 1, and weight scale of 2, inhibitory connection",
				Params: params.Params{
					"Prjn.Learn.Learn": "false",
					"Prjn.Learn.Lrate":  "0",
					"Prjn.WtInit.Dist": "Uniform",
					"Prjn.WtInit.Mean": "1",
					"Prjn.WtInit.Var":  "0",
					"Prjn.WtScale.Abs": "2",	
				}},
				
					
			{Sel: "Layer", Desc: "using 2.3 inhib for all of network -- can explore",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "2.3",
					"Layer.Act.Gbar.L":     "0.1", // set explictly, new default, a bit better vs 0.2
					"Layer.Act.XX1.Gain" : "100",
				}},
			{Sel: "#Behavior", Desc: "Make Behavior layer selective for 1 behavior",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "2.5",
					"Layer.Act.XX1.Gain": "400",
				}},	
			{Sel: "#Approach", Desc: "",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "1.5",
					"Layer.Act.XX1.Gain": "400",
					"Layer.Act.Gbar.E": "1.0",
					"Layer.Act.Noise.Dist": "Gaussian",
					"Layer.Act.Noise.Mean": "0.0",
					"Layer.Act.Noise.Var": "0.0",
					"Layer.Act.Noise.Type": "GeNoise",

				}},	
				
			{Sel: "#Avoid", Desc: "",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "1.3",
					"Layer.Inhib.Layer.FB": "1.0",
					"Layer.Act.XX1.Thr": ".49",
					"Layer.Act.XX1.Gain": "400",
					"Layer.Act.Gbar.E": "1.35",
					"Layer.Act.Noise.Dist": "Gaussian",
					"Layer.Act.Noise.Mean": "0.0",
					"Layer.Act.Noise.Var": "0.0",
					"Layer.Act.Noise.Type": "GeNoise",

				}},	
				
				
			{Sel: "#Hidden1", Desc: "Make Hidden representation a bit sparser",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "2.0",
				}},	
				
			{Sel: "#Hidden2", Desc: "Make Hidden representation a bit sparser",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "2.0",
				}},	

			{Sel: "#VTA", Desc: "Manipulate baseline activation of VTA layer",
				Params: params.Params{
					"Layer.Act.Init.Ge": "0",
				}},	
				
				
			{Sel: ".Back", Desc: "top-down back-projections MUST have lower relative weight scale, otherwise network hallucinates",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.3",
				}},
			
		},
//		"Sim": &params.Sheet{ // sim params apply to sim object
//			{Sel: "Sim", Desc: "best params always finish in this time",
//				Params: params.Params{
//					"Sim.MaxEpcs": "100",
//				}},
//		},
	}},
	{Name: "DefaultInhib", Desc: "output uses default inhib instead of lower", Sheets: params.Sheets{
		"Network": &params.Sheet{
			{Sel: "#Behavior", Desc: "go back to default",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "1.8",
				}},
		},
//		"Sim": &params.Sheet{ // sim params apply to sim object
//			{Sel: "Sim", Desc: "takes longer -- generally doesn't finish..",
//				Params: params.Params{
//					"Sim.MaxEpcs": "100",
//				}},
//		},
	}},
	{Name: "NoMomentum", Desc: "no momentum or normalization", Sheets: params.Sheets{
		"Network": &params.Sheet{
			{Sel: "Prjn", Desc: "no norm or momentum",
				Params: params.Params{
					"Prjn.Learn.Norm.On":     "false",
					"Prjn.Learn.Momentum.On": "false",
				}},
		},
	}},
	{Name: "WtBalOn", Desc: "try with weight bal on", Sheets: params.Sheets{
		"Network": &params.Sheet{
			{Sel: "Prjn", Desc: "weight bal on",
				Params: params.Params{
					"Prjn.Learn.WtBal.On": "true",
				}},
		},
	}},
}
