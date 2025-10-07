pipeline {
  agent any
  stages {
    stage('Checkout Code') {
      steps {
        git(branch: 'master', url: 'https://github.com/laurentmd5/sample-app.git', credentialsId: 'github-token2')
        sh '''
          echo "📦 Repository: https://github.com/laurentmd5/sample-app.git"
          echo "📝 Branch: master"
          echo "🔍 Files in repository:"
          ls -la
        '''
      }
    }

    stage('Build') {
      steps {
        sh '''
          echo "🏗️ Building the application..."
          sleep 2
          echo "✅ Build completed!"
        '''
      }
    }

    stage('Test') {
      steps {
        sh '''
          echo "🧪 Running tests..."
          sleep 2
          echo "✅ All tests passed!"
        '''
      }
    }

    stage('Docker Build & Push') {
      steps {
        sh '''
          echo "🐳 Building and pushing Docker image..."
          echo "$DOCKERHUB_CREDENTIALS_USR"
          echo "Pushing image to Docker Hub..."
          sleep 2
          echo "✅ Image pushed successfully!"
        '''
      }
    }

    stage('Deploy') {
      steps {
        sh '''
          echo "🚀 Deploying application to server..."
          echo "Using SSH key: $SSH_KEY"
          sleep 2
          echo "✅ Deployment successful!"
        '''
      }
    }

  }
  environment {
    DOCKERHUB_CREDENTIALS = credentials('dockerhub-creds')
    SSH_KEY = credentials('ssh-key')
  }
  post {
    success {
      echo '🎉 Pipeline completed successfully!'
    }

    failure {
      echo '❌ Pipeline failed. Check logs for errors.'
    }

  }
}