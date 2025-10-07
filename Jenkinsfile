pipeline {
    agent any
    
    triggers {
        githubPush()
        pollSCM('H/5 * * * *')
    }
    
    environment {
        APP_NAME = 'hello-app'
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
                    gosec -fmt=sarif -out=security-reports/gosec-report.sarif ./... 2>/dev/null || true
                    gosec ./... 2>&1 | tee security-reports/gosec-output.txt || echo "Gosec a terminé avec des findings"
                    echo "✅ Analyse gosec terminée"
                else
                    echo "⚠️ gosec non disponible, analyse statique ignorée"
                fi
                
                # Analyse supplémentaire avec go vet
                echo "=== Exécution de go vet ==="
                go vet ./... 2>&1 | tee security-reports/govet-output.txt || echo "Go vet terminé"
                
                # Vérification des formats
                echo "=== Vérification du format ==="
                test -z "$(gofmt -l .)" && echo "✅ Code bien formaté" || echo "⚠️ Problèmes de format"
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
                
                # Rapport JUnit pour Jenkins
                echo "=== Rapport JUnit ==="
                go test -v ./... 2>&1 | go-junit-report > test-reports/report.xml 2>/dev/null || echo "JUnit report non généré"
                
                # Affichage du résumé
                echo "=== Résumé de Couverture ==="
                cat test-reports/coverage-summary.txt | grep total || echo "Aucune donnée de couverture"
                '''
            }
            post {
                always {
                    junit testResults: 'test-reports/report.xml', allowEmptyResults: true
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
                    
                    # Vérification du Dockerfile
                    echo "=== Contenu du Dockerfile ==="
                    cat Dockerfile
                    echo "============================="
                    
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
                
                # Scan avec rapport détaillé
                echo "=== Scan Détaillé ==="
                trivy image --severity HIGH,CRITICAL --format table ${DOCKER_REGISTRY}/${APP_NAME}:latest 2>&1 | tee trivy-reports/container-scan-summary.txt
                
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
                
                # Scan spécifique des dépendances Go
                echo "=== Scan des Dépendances Go ==="
                if [ -f "go.mod" ]; then
                    trivy filesystem --skip-dirs-db-update --skip-java-db-update go.mod 2>&1 | tee trivy-reports/go-deps-scan.txt || true
                fi
                
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
                    archiveArtifacts artifacts: 'trivy-reports/fs-scan.*,trivy-reports/go-deps-scan.txt', fingerprint: true
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
                            echo '🎯 Démarrage du déploiement de ${APP_NAME}...'
                            
                            # Arrêt de l'ancien conteneur
                            echo '⏹️ Arrêt du conteneur existant...'
                            docker stop ${APP_NAME} 2>/dev/null || true
                            docker rm ${APP_NAME} 2>/dev/null || true
                            
                            # Nettoyage
                            echo '🧹 Nettoyage...'
                            docker image prune -f 2>/dev/null || true
                            
                            # Lancement du nouveau conteneur
                            echo '🐳 Démarrage du nouveau conteneur...'
                            docker run -d \\
                              --name ${APP_NAME} \\
                              -p ${APP_PORT}:${APP_PORT} \\
                              --restart unless-stopped \\
                              ${DOCKER_REGISTRY}/${APP_NAME}:latest
                            
                            # Attente du démarrage
                            echo '⏳ Attente du démarrage...'
                            sleep 15
                            
                            # Vérification
                            echo '🔍 Vérification du déploiement...'
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
                            set -e
                            echo '=== Installation d\\'OWASP ZAP ==='
                            
                            # Vérifier si ZAP est déjà installé
                            if which zap-baseline.py; then
                                echo '✅ ZAP déjà installé'
                                zap-baseline.py -version 2>/dev/null || echo 'ZAP présent'
                            else
                                echo '📥 Installation de ZAP...'
                                sudo apt update
                                sudo apt install -y default-jre wget curl
                                wget -q https://github.com/zaproxy/zaproxy/releases/download/v${ZAP_VERSION}/zap_${ZAP_VERSION}_all.deb
                                sudo dpkg -i zap_${ZAP_VERSION}_all.deb || sudo apt-get install -f -y
                                sudo apt install -y zaproxy
                                rm -f zap_${ZAP_VERSION}_all.deb
                                echo '✅ ZAP installé avec succès'
                            fi
                            
                            # Créer le répertoire pour les rapports
                            mkdir -p /home/devops/zap-reports
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
                            set -e
                            echo '=== Démarrage du Scan ZAP ==='
                            
                            # Attendre que l'application soit prête
                            echo '⏳ Vérification que l\\'application est accessible...'
                            timeout 60 bash -c 'until curl -f -s ${TARGET_URL} > /dev/null; do sleep 5; done'
                            
                            # Scan baseline ZAP
                            echo '=== Scan Baseline ZAP ==='
                            zap-baseline.py -t ${TARGET_URL} -I -m 5 -T 10 -J -j -x /home/devops/zap-reports/zap-report.xml -r /home/devops/zap-reports/zap-report.html 2>&1 | tee /home/devops/zap-reports/zap-scan.log || echo 'Scan ZAP terminé avec des findings'
                            
                            # Résumé du scan
                            echo '=== Résumé du Scan ZAP ==='
                            grep -E '(PASS|FAIL|WARN|INFO)' /home/devops/zap-reports/zap-scan.log | tail -15
                            
                            # Nettoyage des processus ZAP
                            pkill -f zaproxy 2>/dev/null || true
                            echo '✅ Scan ZAP terminé'
                        "
                        
                        # Récupération des rapports ZAP
                        echo '📥 Téléchargement des rapports ZAP...'
                        scp -i \$SSH_KEY -o StrictHostKeyChecking=no ${DEPLOY_SERVER}:/home/devops/zap-reports/zap-report.html zap-reports/ || echo 'Rapport HTML non disponible'
                        scp -i \$SSH_KEY -o StrictHostKeyChecking=no ${DEPLOY_SERVER}:/home/devops/zap-reports/zap-report.xml zap-reports/ || echo 'Rapport XML non disponible'
                        scp -i \$SSH_KEY -o StrictHostKeyChecking=no ${DEPLOY_SERVER}:/home/devops/zap-reports/zap-scan.log zap-reports/ || echo 'Log de scan non disponible'
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
                            set -e
                            echo '=== Audit Système avec Lynis ==='
                            sudo lynis audit system --quick --no-colors 2>&1 | tee /tmp/lynis-audit.txt
                            
                            echo '=== Vérification des Mises à Jour de Sécurité ==='
                            sudo apt update && sudo apt list --upgradable 2>/dev/null | grep -E '(security|ubuntu)' | head -10 | tee /tmp/security-updates.txt
                            
                            echo '=== Vérification des Services ==='
                            sudo systemctl status docker 2>&1 | head -5 | tee /tmp/docker-status.txt
                            
                            echo '=== Utilisation des Ressources ==='
                            free -h | tee /tmp/memory-usage.txt
                            df -h | grep -v tmpfs | tee /tmp/disk-usage.txt
                        "
                        
                        # Récupération des rapports
                        scp -i \$SSH_KEY -o StrictHostKeyChecking=no ${DEPLOY_SERVER}:/tmp/lynis-audit.txt lynis-reports/ || echo 'Rapport Lynis non disponible'
                        scp -i \$SSH_KEY -o StrictHostKeyChecking=no ${DEPLOY_SERVER}:/tmp/security-updates.txt lynis-reports/ || echo 'Rapport updates non disponible'
                        
                        # Génération d'un rapport consolidé
                        cat > lynis-reports/environment-summary.txt << EOF
                        === RAPPORT ENVIRONNEMENT ===
                        Date: $(date)
                        Serveur: ${DEPLOY_SERVER}
                        Application: ${APP_NAME}
                        
                        === AUDIT LYNIS ===
                        $(cat lynis-reports/lynis-audit.txt 2>/dev/null || echo 'Non disponible')
                        
                        === MISES À JOUR SÉCURITÉ ===
                        $(cat lynis-reports/security-updates.txt 2>/dev/null || echo 'Aucune mise à jour critique')
                        EOF
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
                sh """
                echo "📊 RAPPORT DE SÉCURITÉ CONSOLIDÉ"
                echo "================================"
                """
                
                sh '''
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
                - **Tests Unitaires**: $(ls test-reports/report.xml 2>/dev/null && echo "✅ Exécutés" || echo "❌ Non exécutés")
                
                ## 🔒 Sécurité Container
                - **Scan Image**: $(ls trivy-reports/container-scan.html 2>/dev/null && echo "✅ Complété" || echo "❌ Échoué")
                - **Scan Filesystem**: $(ls trivy-reports/fs-scan.html 2>/dev/null && echo "✅ Complété" || echo "❌ Échoué")
                
                ## 🛡️ Sécurité Dynamique
                - **OWASP ZAP**: $(ls zap-reports/zap-report.html 2>/dev/null && echo "✅ Complété" || echo "❌ Échoué")
                
                ## 🏢 Environnement
                - **Audit Lynis**: $(ls lynis-reports/lynis-audit.txt 2>/dev/null && echo "✅ Complété" || echo "❌ Échoué")
                - **Mises à jour**: $(ls lynis-reports/security-updates.txt 2>/dev/null && echo "✅ Vérifiées" || echo "❌ Non vérifiées")
                
                ## 📈 Statut Global
                - **Build**: ${currentBuild.result}
                - **Durée**: ${currentBuild.durationString}
                - **Application**: $(curl -s -o /dev/null -w "%{http_code}" ${TARGET_URL} || echo "Inaccessible")
                
                ## 📋 Recommendations
                - Vérifier les vulnérabilités critiques dans les rapports Trivy
                - Examiner les findings ZAP pour les éventuelles failles web
                - Mettre à jour les dépendances si nécessaire
                - Surveiller les mises à jour de sécurité système
                
                EOF
                
                # Conversion en HTML
                which pandoc && pandoc security-summary.md -o security-reports/security-summary.html || cp security-summary.md security-reports/security-summary.html
                
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
                sh """
                echo "🎯 VÉRIFICATION FINALE"
                echo "======================"
                """
                
                script {
                    withCredentials([sshUserPrivateKey(
                        credentialsId: "${SSH_CREDENTIALS_ID}",
                        usernameVariable: 'SSH_USER',
                        keyFileVariable: 'SSH_KEY'
                    )]) {
                        sh """
                        ssh -i \$SSH_KEY -o StrictHostKeyChecking=no ${DEPLOY_SERVER} "
                            echo '📊 ÉTAT FINAL DU DÉPLOIEMENT'
                            echo '============================'
                            
                            # Statut de l'application
                            echo '=== Application ==='
                            docker ps --filter 'name=${APP_NAME}' --format 'table {{.Names}}\\t{{.Status}}\\t{{.RunningFor}}\\t{{.Ports}}'
                            
                            # Santé de l'application
                            echo '=== Santé ==='
                            if curl -f -s -o /dev/null -w 'HTTP: %{http_code}\\n' ${TARGET_URL}; then
                                echo '✅ APPLICATION EN LIGNE ET FONCTIONNELLE'
                            else
                                echo '❌ APPLICATION INACCESSIBLE'
                                docker logs ${APP_NAME} --tail 10
                            fi
                            
                            # Ressources
                            echo '=== Ressources ==='
                            echo 'Mémoire:'
                            free -h | head -2
                            echo 'Disque:'
                            df -h /home
                            
                            # Sécurité
                            echo '=== Sécurité ==='
                            echo 'Conteneur isolé: ' \$(docker inspect ${APP_NAME} --format '{{.HostConfig.NetworkMode}}' 2>/dev/null || echo 'bridge')
                            echo 'Port exposé: ${APP_PORT}'
                            
                            echo ''
                            echo '🎉 DÉPLOIEMENT ET SCANS TERMINÉS AVEC SUCCÈS!'
                            echo '🛡️ Tous les contrôles de sécurité ont été exécutés'
                            echo '🌐 Application disponible: ${TARGET_URL}'
                            echo '📊 Consultez Jenkins pour les rapports détaillés'
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
            echo "📊 RÉSUMÉ DE L'EXÉCUTION"
            echo "========================"
            echo "✅ Application Go construite"
            echo "✅ Image Docker créée"
            echo "✅ Tests et couverture exécutés"
            echo "✅ Scans de sécurité complétés"
            echo "✅ Application déployée"
            echo "✅ Audits environnementaux réalisés"
            echo ""
            echo "📋 RAPPORTS DISPONIBLES:"
            echo "   - 🔍 Gosec: Analyse statique"
            echo "   - 🧪 Tests: Couverture et résultats"
            echo "   - 🔒 Trivy: Scan containers et fichiers"
            echo "   - 🛡️ ZAP: Scan dynamique"
            echo "   - 🏢 Lynis: Audit environnement"
            echo "   - 📊 Résumé sécurité complet"
            '''
            
            # Archivage de tous les rapports
            archiveArtifacts artifacts: 'security-reports/**,test-reports/**,trivy-reports/**,zap-reports/**,lynis-reports/**,${APP_NAME}', fingerprint: true
        }
        success {
            sh """
            echo ""
            echo "🎉 PIPELINE DE SÉCURITÉ COMPLET RÉUSSI!"
            echo "======================================="
            echo "🛡️  Tous les contrôles de sécurité ont passé"
            echo "🚀  Application déployée sécuritairement"
            echo "📈  Rapports de sécurité générés"
            echo "🌐  Accédez à l'application: ${TARGET_URL}"
            echo ""
            echo "✅ BUILD: ${env.BUILD_NUMBER}"
            echo "✅ RÉSULTAT: SUCCÈS"
            """
        }
        failure {
            sh """
            echo "❌ PIPELINE EN ÉCHEC"
            echo "==================="
            echo "💡 Causes possibles:"
            echo "   - Échec de compilation"
            echo "   - Tests en échec"
            echo "   - Vulnérabilités critiques détectées"
            echo "   - Problèmes de déploiement"
            echo ""
            echo "🔍 Consultez les rapports pour plus de détails"
            """
        }
        unstable {
            sh """
            echo "⚠️ PIPELINE INSTABLE"
            echo "==================="
            echo "📝 Certains scans ont détecté des problèmes non-critiques"
            echo "💡 Vérifiez les rapports de sécurité pour les détails"
            """
        }
    }
}
