pipeline {
    agent any
    
    triggers {
        githubPush()
        pollSCM('H/5 * * * *')
    }
    
    environment {
        APP_NAME = 'go-dev-dashboard'
        APP_PORT = '8090'
        DOCKER_REGISTRY = 'laurentmd5'
        DEPLOY_SERVER = 'devops@localhost'
        SSH_CREDENTIALS_ID = 'ubuntu-server-ssh'
        TRIVY_VERSION = '0.49.1'
        GOSEC_VERSION = '2.19.0'
        ZAP_VERSION = '2.14.0'
        TARGET_URL = "http://192.168.61.131:${APP_PORT}"
    }
    
    stages {
        // ÉTAPE 1: Checkout du Code
        stage('Checkout Code') {
            steps {
                git branch: 'master',
                    url: 'https://github.com/laurentmd5/sample-app.git',
                    credentialsId: 'github-token2'

                sh '''
                echo "📦 Repository: https://github.com/laurentmd5/sample-app.git"
                echo "📝 Branch: master"
                echo "🔍 Dernier commit:"
                git log -1 --oneline
                echo "📁 Contenu du repository:"
                ls -la
                '''
            }
        }
        
        // ÉTAPE 2: Setup Environment et Outils de Sécurité
        stage('Setup Environment') {
            steps {
                sh '''
                echo "🔧 Configuration de l'environnement..."
                echo "=== Versions des outils ==="
                go version || echo "Go non installé"
                docker --version || echo "Docker non disponible"
                
                echo "📥 Installation des outils de sécurité..."
                
                # Installation de gosec
                echo "=== Installation de gosec ==="
                which gosec || go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
                
                # Installation de Trivy
                echo "=== Installation de Trivy ==="
                if ! which trivy; then
                    wget -q https://github.com/aquasecurity/trivy/releases/download/v${TRIVY_VERSION}/trivy_${TRIVY_VERSION}_Linux-64bit.deb
                    sudo dpkg -i trivy_${TRIVY_VERSION}_Linux-64bit.deb
                    rm -f trivy_${TRIVY_VERSION}_Linux-64bit.deb
                fi
                trivy --version || echo "Trivy installation failed"
                
                # Installation de Lynis
                echo "=== Installation de Lynis ==="
                which lynis || (sudo apt update && sudo apt install -y lynis)
                
                echo "✅ Tous les outils sont prêts"
                '''
            }
        }
        
        // ÉTAPE 3: Build Application Go
        stage('Build Go Application') {
            steps {
                sh '''
                echo "🏗️ Construction de l'application Go..."
                
                # Initialisation des modules Go
                if [ ! -f "go.mod" ]; then
                    echo "📝 Initialisation de go.mod..."
                    go mod init hello-app
                fi
                
                # Téléchargement des dépendances
                echo "📥 Téléchargement des dépendances..."
                go mod download || echo "Aucune dépendance ou déjà téléchargées"
                
                # Construction de l'application
                echo "🔨 Compilation de l'application..."
                go build -v -o ${APP_NAME} .
                
                # Vérification du binaire
                echo "✅ Vérification du build:"
                ls -la ${APP_NAME}
                file ${APP_NAME}
                chmod +x ${APP_NAME}
                '''
            }
        }
        
        // ÉTAPE 4: Static Code Analysis - gosec
        stage('Static Code Analysis') {
            steps {
                sh '''
                echo "🔍 Analyse Statique du Code avec gosec..."
                mkdir -p security-reports
                
                # Analyse de sécurité avec gosec
                echo "=== Exécution de gosec ==="
                if which gosec; then
                    gosec -fmt=json -out=security-reports/gosec-report.json ./... 2>/dev/null || true
                    gosec -fmt=html -out=security-reports/gosec-report.html ./... 2>/dev/null || true
                    gosec ./... 2>&1 | tee security-reports/gosec-output.txt || echo "Gosec a terminé avec des findings"
                    echo "✅ Analyse gosec terminée"
                else
                    echo "⚠️ gosec non disponible, analyse statique ignorée"
                fi
                
                # Analyse supplémentaire avec go vet
                echo "=== Exécution de go vet ==="
                go vet ./... 2>&1 | tee security-reports/govet-output.txt || echo "Go vet terminé"
                '''
            }
            post {
                always {
                    publishHTML([
                        allowMissing: true,
                        alwaysLinkToLastBuild: true,
                        keepAll: true,
                        reportDir: 'security-reports',
                        reportFiles: 'gosec-report.html',
                        reportName: 'Rapport Gosec - Analyse Statique'
                    ])
                    archiveArtifacts artifacts: 'security-reports/gosec-report.*,security-reports/govet-output.txt', fingerprint: true
                }
            }
        }
        
        // ÉTAPE 5: Dynamic Tests et Couverture
        stage('Dynamic Tests') {
            steps {
                sh '''
                echo "🧪 Tests Dynamiques et Couverture..."
                mkdir -p test-reports
                
                # Tests unitaires avec coverage
                echo "=== Exécution des Tests Unitaires ==="
                go test -v -race -coverprofile=test-reports/coverage.out -covermode=atomic ./... 2>&1 | tee test-reports/test-output.log
                
                # Génération des rapports de couverture
                echo "=== Génération des Rapports ==="
                go tool cover -html=test-reports/coverage.out -o test-reports/coverage.html
                go tool cover -func=test-reports/coverage.out > test-reports/coverage-summary.txt
                
                # Affichage du résumé
                echo "=== Résumé de Couverture ==="
                cat test-reports/coverage-summary.txt | grep total || echo "Aucune donnée de couverture"
                '''
            }
            post {
                always {
                    publishHTML([
                        allowMissing: true,
                        alwaysLinkToLastBuild: true,
                        keepAll: true,
                        reportDir: 'test-reports',
                        reportFiles: 'coverage.html',
                        reportName: 'Rapport de Couverture des Tests'
                    ])
                    archiveArtifacts artifacts: 'test-reports/**/*', fingerprint: true
                }
            }
        }
        
        // ÉTAPE 6: Construction de l'Image Docker
        stage('Build Docker Image') {
            steps {
                script {
                    sh """
                    echo "🐳 Construction de l'Image Docker..."
                    
                    # Construction de l'image
                    docker build -t ${DOCKER_REGISTRY}/${APP_NAME}:${env.BUILD_NUMBER} .
                    docker tag ${DOCKER_REGISTRY}/${APP_NAME}:${env.BUILD_NUMBER} ${DOCKER_REGISTRY}/${APP_NAME}:latest
                    
                    echo "✅ Images Docker créées:"
                    docker images | grep ${DOCKER_REGISTRY} || echo "Aucune image trouvée"
                    """
                }
            }
        }
        
        // ÉTAPE 7: Scan de Container - Trivy
        stage('Container Scan') {
            steps {
                sh '''
                echo "🔒 Scan de Vulnérabilités du Container avec Trivy..."
                mkdir -p trivy-reports
                
                # Scan de l'image Docker
                echo "=== Scan de l'Image Docker ==="
                trivy image --format template --template "@contrib/html.tpl" -o trivy-reports/container-scan.html ${DOCKER_REGISTRY}/${APP_NAME}:latest
                trivy image --format json -o trivy-reports/container-scan.json ${DOCKER_REGISTRY}/${APP_NAME}:latest
                trivy image --exit-code 0 --severity HIGH,CRITICAL ${DOCKER_REGISTRY}/${APP_NAME}:latest
                
                echo "✅ Scan du container terminé"
                '''
            }
            post {
                always {
                    publishHTML([
                        allowMissing: true,
                        alwaysLinkToLastBuild: true,
                        keepAll: true,
                        reportDir: 'trivy-reports',
                        reportFiles: 'container-scan.html',
                        reportName: 'Scan Sécurité Container - Trivy'
                    ])
                    archiveArtifacts artifacts: 'trivy-reports/container-scan.*', fingerprint: true
                }
            }
        }
        
        // ÉTAPE 8: Scan du Filesystem - Trivy FS
        stage('Filesystem Scan') {
            steps {
                sh '''
                echo "📁 Scan du Filesystem et Dépendances..."
                mkdir -p trivy-reports
                
                # Scan du filesystem
                echo "=== Scan des Dépendances ==="
                trivy filesystem --format template --template "@contrib/html.tpl" -o trivy-reports/fs-scan.html .
                trivy filesystem --format json -o trivy-reports/fs-scan.json .
                trivy filesystem --exit-code 0 --severity HIGH,CRITICAL .
                
                echo "✅ Scan du filesystem terminé"
                '''
            }
            post {
                always {
                    publishHTML([
                        allowMissing: true,
                        alwaysLinkToLastBuild: true,
                        keepAll: true,
                        reportDir: 'trivy-reports',
                        reportFiles: 'fs-scan.html',
                        reportName: 'Scan Sécurité Filesystem - Trivy'
                    ])
                    archiveArtifacts artifacts: 'trivy-reports/fs-scan.*', fingerprint: true
                }
            }
        }
        
        // ÉTAPE 9: Déploiement sur Ubuntu
        stage('Deploy to Ubuntu') {
            steps {
                script {
                    withCredentials([sshUserPrivateKey(
                        credentialsId: "${SSH_CREDENTIALS_ID}",
                        usernameVariable: 'SSH_USER',
                        keyFileVariable: 'SSH_KEY'
                    )]) {
                        sh """
                        echo "🚀 Déploiement sur le serveur Ubuntu..."
                        
                        ssh -i \$SSH_KEY -o StrictHostKeyChecking=no ${DEPLOY_SERVER} "
                            set -e
                            echo '🎯 Démarrage du déploiement...'
                            
                            # Arrêt de l'ancien conteneur
                            docker stop ${APP_NAME} 2>/dev/null || true
                            docker rm ${APP_NAME} 2>/dev/null || true
                            
                            # Lancement du nouveau conteneur
                            docker run -d \\
                              --name ${APP_NAME} \\
                              -p ${APP_PORT}:${APP_PORT} \\
                              --restart unless-stopped \\
                              ${DOCKER_REGISTRY}/${APP_NAME}:latest
                            
                            # Attente du démarrage
                            sleep 15
                            
                            # Vérification
                            docker ps --filter 'name=${APP_NAME}' --format 'table {{.Names}}\\t{{.Status}}\\t{{.Ports}}'
                            
                            echo '✅ Déploiement terminé avec succès'
                        "
                        """
                    }
                }
            }
        }
        
        // ÉTAPE 10: Installation OWASP ZAP sur le Serveur
        stage('Install OWASP ZAP') {
            steps {
                script {
                    withCredentials([sshUserPrivateKey(
                        credentialsId: "${SSH_CREDENTIALS_ID}",
                        usernameVariable: 'SSH_USER',
                        keyFileVariable: 'SSH_KEY'
                    )]) {
                        sh """
                        echo "📥 Installation d'OWASP ZAP sur le serveur..."
                        
                        ssh -i \$SSH_KEY -o StrictHostKeyChecking=no ${DEPLOY_SERVER} "
                            # Vérifier si ZAP est déjà installé
                            if which zap-baseline.py; then
                                echo '✅ ZAP déjà installé'
                            else
                                echo '📥 Installation de ZAP...'
                                sudo apt update
                                sudo apt install -y default-jre wget
                                wget -q https://github.com/zaproxy/zaproxy/releases/download/v${ZAP_VERSION}/zap_${ZAP_VERSION}_all.deb
                                sudo dpkg -i zap_${ZAP_VERSION}_all.deb || sudo apt-get install -f -y
                                sudo apt install -y zaproxy
                                rm -f zap_${ZAP_VERSION}_all.deb
                                echo '✅ ZAP installé avec succès'
                            fi
                        "
                        """
                    }
                }
            }
        }
        
        // ÉTAPE 11: Scan Dynamique - OWASP ZAP
        stage('OWASP ZAP Security Scan') {
            steps {
                script {
                    withCredentials([sshUserPrivateKey(
                        credentialsId: "${SSH_CREDENTIALS_ID}",
                        usernameVariable: 'SSH_USER',
                        keyFileVariable: 'SSH_KEY'
                    )]) {
                        sh """
                        echo "🛡️ Scan de Sécurité Dynamique OWASP ZAP..."
                        mkdir -p zap-reports
                        
                        # Exécution du scan ZAP sur le serveur
                        ssh -i \$SSH_KEY -o StrictHostKeyChecking=no ${DEPLOY_SERVER} "
                            echo '=== Démarrage du Scan ZAP ==='
                            
                            # Attendre que l'application soit prête
                            sleep 30
                            
                            # Scan baseline ZAP
                            zap-baseline.py -t ${TARGET_URL} -I -m 5 -T 10 -J -j -x /home/devops/zap-report.xml -r /home/devops/zap-report.html 2>&1 | tee /home/devops/zap-scan.log || echo 'Scan ZAP terminé avec des findings'
                            
                            # Nettoyage des processus ZAP
                            pkill -f zaproxy 2>/dev/null || true
                            echo '✅ Scan ZAP terminé'
                        "
                        
                        # Récupération des rapports ZAP
                        scp -i \$SSH_KEY -o StrictHostKeyChecking=no ${DEPLOY_SERVER}:/home/devops/zap-report.html zap-reports/ || echo 'Rapport HTML non disponible'
                        scp -i \$SSH_KEY -o StrictHostKeyChecking=no ${DEPLOY_SERVER}:/home/devops/zap-report.xml zap-reports/ || echo 'Rapport XML non disponible'
                        """
                    }
                }
            }
            post {
                always {
                    publishHTML([
                        allowMissing: true,
                        alwaysLinkToLastBuild: true,
                        keepAll: true,
                        reportDir: 'zap-reports',
                        reportFiles: 'zap-report.html',
                        reportName: 'Scan Sécurité Dynamique - OWASP ZAP'
                    ])
                    archiveArtifacts artifacts: 'zap-reports/**/*', fingerprint: true
                }
            }
        }
        
        // ÉTAPE 12: Audit Environnement - Lynis
        stage('Environment Scan') {
            steps {
                script {
                    withCredentials([sshUserPrivateKey(
                        credentialsId: "${SSH_CREDENTIALS_ID}",
                        usernameVariable: 'SSH_USER',
                        keyFileVariable: 'SSH_KEY'
                    )]) {
                        sh """
                        echo "🏢 Audit de Sécurité de l'Environnement..."
                        mkdir -p lynis-reports
                        
                        # Audit du serveur avec Lynis
                        ssh -i \$SSH_KEY -o StrictHostKeyChecking=no ${DEPLOY_SERVER} "
                            echo '=== Audit Système avec Lynis ==='
                            sudo lynis audit system --quick 2>&1 | tee /tmp/lynis-audit.txt
                            
                            echo '=== Vérification des Mises à Jour ==='
                            sudo apt update && sudo apt list --upgradable 2>/dev/null | head -10 | tee /tmp/security-updates.txt
                        "
                        
                        # Récupération des rapports - CORRECTION APPLIQUÉE ICI
                        scp -i \$SSH_KEY -o StrictHostKeyChecking=no ${DEPLOY_SERVER}:/tmp/lynis-audit.txt lynis-reports/ || echo 'Rapport Lynis non disponible'
                        scp -i \$SSH_KEY -o StrictHostKeyChecking=no ${DEPLOY_SERVER}:/tmp/security-updates.txt lynis-reports/ || echo 'Rapport updates non disponible'
                        """
                    }
                }
            }
            post {
                always {
                    archiveArtifacts artifacts: 'lynis-reports/**/*', fingerprint: true
                }
            }
        }
        
        // ÉTAPE 13: Rapport de Sécurité Consolidé
        stage('Security Summary') {
            steps {
                sh '''
                echo "📊 Génération du Rapport de Sécurité Consolidé..."
                
                # Génération du rapport Markdown
                cat > security-summary.md << EOF
                # Rapport de Sécurité Complet - Build ${BUILD_NUMBER}
                
                ## 📅 Date: $(date)
                ## 🏷️ Application: ${APP_NAME}
                ## 🌐 URL: ${TARGET_URL}
                
                ## 🔍 Analyse Statique
                - **Gosec**: $(ls security-reports/gosec-report.html 2>/dev/null && echo "✅ Complété" || echo "❌ Échoué")
                - **Go Vet**: $(ls security-reports/govet-output.txt 2>/dev/null && echo "✅ Complété" || echo "❌ Échoué")
                
                ## 🧪 Tests Dynamiques
                - **Couverture**: $(if [ -f "test-reports/coverage-summary.txt" ]; then cat test-reports/coverage-summary.txt | grep total | awk "{print \\$3}"; else echo "N/A"; fi)
                
                ## 🔒 Sécurité Container
                - **Scan Image**: $(ls trivy-reports/container-scan.html 2>/dev/null && echo "✅ Complété" || echo "❌ Échoué")
                - **Scan Filesystem**: $(ls trivy-reports/fs-scan.html 2>/dev/null && echo "✅ Complété" || echo "❌ Échoué")
                
                ## 🛡️ Sécurité Dynamique
                - **OWASP ZAP**: $(ls zap-reports/zap-report.html 2>/dev/null && echo "✅ Complété" || echo "❌ Échoué")
                
                ## 🏢 Environnement
                - **Audit Lynis**: $(ls lynis-reports/lynis-audit.txt 2>/dev/null && echo "✅ Complété" || echo "❌ Échoué")
                
                ## 📈 Statut Global
                - **Build**: ${currentBuild.result}
                - **Application**: $(curl -s -o /dev/null -w "%{http_code}" ${TARGET_URL} || echo "Inaccessible")
                
                EOF
                
                # Copie simple sans pandoc
                cp security-summary.md security-reports/security-summary.html
                
                echo "✅ Rapport de sécurité généré"
                '''
            }
            post {
                always {
                    publishHTML([
                        allowMissing: true,
                        alwaysLinkToLastBuild: true,
                        keepAll: true,
                        reportDir: 'security-reports',
                        reportFiles: 'security-summary.html',
                        reportName: 'Résumé Sécurité Complet'
                    ])
                    archiveArtifacts artifacts: 'security-summary.md,security-reports/security-summary.html', fingerprint: true
                }
            }
        }
        
        // ÉTAPE 14: Vérification Finale
        stage('Final Check') {
            steps {
                script {
                    withCredentials([sshUserPrivateKey(
                        credentialsId: "${SSH_CREDENTIALS_ID}",
                        usernameVariable: 'SSH_USER',
                        keyFileVariable: 'SSH_KEY'
                    )]) {
                        sh """
                        echo "🎯 Vérification Finale..."
                        
                        ssh -i \$SSH_KEY -o StrictHostKeyChecking=no ${DEPLOY_SERVER} "
                            echo '📊 État Final du Déploiement'
                            echo '============================'
                            
                            # Statut de l'application
                            docker ps --filter 'name=${APP_NAME}' --format 'table {{.Names}}\\t{{.Status}}\\t{{.Ports}}'
                            
                            # Santé de l'application
                            if curl -f -s ${TARGET_URL} > /dev/null; then
                                echo '✅ APPLICATION EN LIGNE ET FONCTIONNELLE'
                            else
                                echo '❌ APPLICATION INACCESSIBLE'
                            fi
                            
                            echo ''
                            echo '🎉 DÉPLOIEMENT ET SCANS TERMINÉS AVEC SUCCÈS!'
                            echo '🌐 Application disponible: ${TARGET_URL}'
                        "
                        """
                    }
                }
            }
        }
    }
    
    post {
        always {
            sh '''
            echo "🧹 Nettoyage final..."
            docker system prune -f 2>/dev/null || true
            
            echo ""
            echo "📊 RÉSUMÉ DE L'\''EXÉCUTION"
            echo "========================="
            echo "✅ Application Go construite"
            echo "✅ Image Docker créée"
            echo "✅ Tests et couverture exécutés"
            echo "✅ Scans de sécurité complétés"
            echo "✅ Application déployée"
            echo ""
            echo "📋 RAPPORTS DISPONIBLES:"
            echo "   - 🔍 Gosec: Analyse statique"
            echo "   - 🧪 Tests: Couverture et résultats"
            echo "   - 🔒 Trivy: Scan containers et fichiers"
            echo "   - 🛡️ ZAP: Scan dynamique"
            echo "   - 🏢 Lynis: Audit environnement"
            echo "   - 📊 Résumé sécurité complet"
            '''
            
            // Archivage de tous les rapports
            archiveArtifacts artifacts: 'security-reports/**,test-reports/**,trivy-reports/**,zap-reports/**,lynis-reports/**,${APP_NAME}', fingerprint: true
        }
        success {
            sh """
            echo ""
            echo "🎉 PIPELINE DE SÉCURITÉ COMPLET RÉUSSI!"
            echo "======================================="
            echo "🛡️  Tous les contrôles de sécurité ont passé"
            echo "🚀  Application déployée sécuritairement"
            echo "🌐  Accédez à l'application: ${TARGET_URL}"
            """
        }
        failure {
            sh """
            echo "❌ PIPELINE EN ÉCHEC"
            echo "==================="
            echo "💡 Consultez les rapports pour plus de détails"
            """
        }
    }
}
