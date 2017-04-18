node('golang') {
	def root = tool name: 'Go 1.8.1', type: 'go'
	def version = '0.1.5'

	checkout([$class: 'GitSCM',
		branches: [[name: '*/master']], 
		doGenerateSubmoduleConfigurations: false, 
		extensions: [[$class: 'RelativeTargetDirectory', relativeTargetDir: 'src/github.com/csduarte/mattermost-probe']], 
		submoduleCfg: [], 
		userRemoteConfigs: [[credentialsId: 'git-key',url: 'git@github.com:uberdeploy/mattermost-probe.git']]])

	stage('prep') {
		withEnv(["GOROOT=${root}", "GOPATH=${WORKSPACE}", "PATH+GO=${root}/bin"]) {
			sh 'go version'
			sh 'cd $WORKSPACE/src/github.com/csduarte/mattermost-probe && glide install'
			sh 'if [[ ! -d $WORKSPACE/bin ]]; then mkdir $WORKSPACE/bin; fi; if [[ ! -d $WORKSPACE/pkg ]]; then mkdir $WORKSPACE/pkg; fi'
		}
	}

	stage('test') {
		withEnv(["GOROOT=${root}", "GOPATH=${WORKSPACE}", "PATH+GO=${root}/bin"]) {
			sh 'cd $WORKSPACE && go test github.com/csduarte/mattermost-probe/mattermost'
		}
	}

	stage('build') {
		withEnv(["GOROOT=${root}", "GOPATH=${WORKSPACE}", "PATH+GO=${root}/bin"]) {
			sh 'cd $WORKSPACE && go build github.com/csduarte/mattermost-probe'
		}
	}

	stage('publish') {
		nexusArtifactUploader artifacts: [[ artifactId: 'mattermost-probe',
											classifier: '',
											file: 'mattermost-probe',
											type: 'bin' ]],
							  credentialsId: 'nexus-deploy',
							  groupId: 'com.github.uberdeploy.mattermost-probe',
							  nexusUrl: 'nexus.test',
							  nexusVersion: 'nexus3',
							  protocol: 'https',
							  repository: 'maven-releases',
							  version: version
	}
}
