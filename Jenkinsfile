pipeline {
  agent {
    label 'docker-migration'
  }
  options {
    timestamps()
  }
  environment {
    HELLO = 'world'
  }
  stages {
    stage('Root CA') {
        sh 'make letsencrypt-root'
    }
    stage('Build Linux') {
      environment {
        GOOS='linux'
        GOARCH='amd64'
      }
      steps {
        lock(resource: "${JOB_NAME}-linux") {
          milestone(ordinal: 20, label: 'linux milestone')
          sh 'make release-server'
          sh 'make release-client'
        }
      }
    }
    stage('Build Darwin') {
      environment {
        GOOS='darwin'
        GOARCH='amd64'
      }
      steps {
        lock(resource: "${job_name}-darwin") {
          echo "obtained lock: ${job_name}-build-darwin"
          milestone(ordinal: 30, label: 'build darwin milestone')
          sh 'make release-server'
          sh 'make release-client'
          archiveartifacts artifacts: 'bin/darwin_amd64/ngrokd', fingerprint: true
          archiveartifacts artifacts: 'bin/darwin_amd64/ngrok', fingerprint: true
        }
      }
    }
    stage('Build Docker') {
      steps {
         lock(resource: "${job_name}-build-docker") {
          echo "obtained lock: ${job_name}-build-docker"
          milestone(ordinal: 40, label: 'build docker')
          sh 'make docker-image'
        }
      }
    }
    stage('Sonarqube') {
      steps {
        lock(resource: "${JOB_NAME}-sonarqube") {
          echo "obtained lock: ${JOB_NAME}-sonarqube"
          milestone(ordinal: 20, label: 'sonarqube milestone')
          sh 'make docker-sonar'
        }
      }
    }
    stage('Tag') {
      when {
        expression { BRANCH_NAME ==~ /master/ }
      }
      steps {
        milestone(ordinal: 30, label: 'docker tag milestone')
        sh 'make docker-tag'
      }
    }
    stage('Push') {
      when {
        expression { BRANCH_NAME ==~ /master/ }
      }
      steps {
        milestone(ordinal: 40, label: 'docker push milestone')
        sh '''
          eval "$(aws ecr get-login --region "$AWS_ECR_REGION" --registry-ids "$AWS_ECR_REGISTRY_ID" | sed 's/-e none //')"
        '''
        sh 'make ecr-repository'
        sh 'make docker-push'
      }
    }
  }
  post {
    failure {
      slackSend(color: '#E7433E', message: "FAILED: ${JOB_NAME} #${BUILD_NUMBER} (<${BUILD_URL}|Open>)")
    }
    aborted {
      slackSend(color: '#930DFF', message: "ABORTED: ${JOB_NAME} #${BUILD_NUMBER} (<${BUILD_URL}|Open>)")
    }
  }
}
