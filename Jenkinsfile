node('go1.17') {
	def newTag = ''
	def tag = ''
	def gitTag = ''

	try {
		stage('Checkout'){
				checkout scm
				notifyBitbucket()
				gitTag = sh(script: 'git tag -l --contains HEAD', returnStdout: true).trim()
		}


		stage('Fetch dependencies'){
			sh('dep ensure -v -vendor-only')
		}
		stage('Run test'){
			sh('go test -v ./...')
		}

		if(gitTag != ''){
			tag = gitTag
		}

		if( tag != ''){
			strippedTag = tag.replaceFirst('v', '')
			stage('Build the application'){
				echo "Building with docker tag ${strippedTag}"
				sh('CGO_ENABLED=0 GOOS=linux go build')
			}

			stage('Generate docker image'){
				image = docker.build('fortnox/alertmanager-bot:'+strippedTag, '--pull .')
			}

			stage('Push docker image'){
				docker.withRegistry("https://quay.io", 'docker-registry') {
					image.push()
				}
			}
		}

		currentBuild.result = 'SUCCESS'
	} catch(err) {
		currentBuild.result = 'FAILED'
		notifyBitbucket()
		throw err
	}

	notifyBitbucket()
}

