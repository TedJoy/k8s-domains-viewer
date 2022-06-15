package k8s

import (
	appcfg "git2.gnt-global.com/jlab/gdeploy/domains-viewers/config"
	"go.uber.org/zap"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// var (
// 	universalDeserializer = serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()
// 	// clientSet             *kubernetes.Clientset
// )

func GetClientSet(logger *zap.SugaredLogger) *kubernetes.Clientset {
	var config *rest.Config
	if !appcfg.MyEnvConfig.Application.UseKubeCfg {
		// default to service account in cluster token
		c, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		config = c
	} else {
		kubeconfig := appcfg.MyEnvConfig.Application.KubeConfigFile

		logger.Debug("kubeconfig: " + kubeconfig)

		c, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}
		config = c
		kct, err := GetKubeContext(&kubeconfig)
		if err != nil {
			panic(err.Error())
		}
		logger.Debug("kubecontext: " + kct)
	}

	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return cs
}

func GetDynamic(logger *zap.SugaredLogger) dynamic.Interface {
	var config *rest.Config
	if !appcfg.MyEnvConfig.Application.UseKubeCfg {
		// default to service account in cluster token
		c, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		config = c
	} else {
		kubeconfig := appcfg.MyEnvConfig.Application.KubeConfigFile

		logger.Debug("kubeconfig: " + kubeconfig)

		c, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}
		config = c
		kct, err := GetKubeContext(&kubeconfig)
		if err != nil {
			panic(err.Error())
		}
		logger.Debug("kubecontext: " + kct)
	}

	cs, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return cs
}

func GetKubeContext(pathToKubeConfig *string) (string, error) {
	rawconfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: *pathToKubeConfig},
		&clientcmd.ConfigOverrides{
			CurrentContext: "",
		}).RawConfig()
	return rawconfig.CurrentContext, err
}
