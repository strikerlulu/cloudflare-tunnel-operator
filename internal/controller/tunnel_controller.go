/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"

	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/option"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
	networkingv1alpha1 "github.com/strikerlulu/cloudflare-tunnel-operator/api/v1alpha1"
)

type TunnelReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const tunnelFinalizer = "networking.strikerlulu.me/finalizer"

func (r *TunnelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Starting reconciliation")

	tunnel := &networkingv1alpha1.Tunnel{}
	if err := r.Get(ctx, req.NamespacedName, tunnel); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("Tunnel resource found", "Name", tunnel.Name, "Namespace", tunnel.Namespace)

	secretNamespace := tunnel.Namespace
	if tunnel.Spec.SecretNamespace != "" {
		secretNamespace = tunnel.Spec.SecretNamespace
	}

	log.Info("Fetching secret", "Name", tunnel.Spec.SecretRef, "Namespace", secretNamespace)

	secret := &corev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{Name: tunnel.Spec.SecretRef, Namespace: secretNamespace}, secret)
	if err != nil {
		log.Error(err, "Failed to get Cloudflare Secret")
		return ctrl.Result{RequeueAfter: time.Minute * 1}, nil
	}
	apiToken := string(secret.Data["api-token"])

	cf := cloudflare.NewClient(option.WithAPIToken(apiToken))

	log.Info("Checking deletion timestamp")

	if !tunnel.ObjectMeta.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(tunnel, tunnelFinalizer) {
			log.Info("Tunnel marked for deletion, removing configuration")

			tunnels, err := cf.ZeroTrust.Tunnels.List(ctx, zero_trust.TunnelListParams{
				AccountID: cloudflare.F(tunnel.Spec.AccountID),
			})
			if err != nil {
				log.Error(err, "Failed to list tunnels for cleanup")
				return ctrl.Result{RequeueAfter: time.Minute * 1}, nil
			}

			var tunnelID string
			for _, t := range tunnels.Result {
				if t.Name == tunnel.Spec.SharedTunnelName {
					tunnelID = t.ID
					break
				}
			}

			if tunnelID != "" {
				config, err := cf.ZeroTrust.Tunnels.Cloudflared.Configurations.Get(ctx, tunnelID, zero_trust.TunnelCloudflaredConfigurationGetParams{
					AccountID: cloudflare.F(tunnel.Spec.AccountID),
				})
				if err == nil {
					// Filter out the rule to be removed
					var newIngress []zero_trust.TunnelCloudflaredConfigurationUpdateParamsConfigIngress
					for _, ing := range config.Config.Ingress {
						// Only keep rules that are NOT the target domain AND are NOT the catch-all
						if ing.Hostname != tunnel.Spec.Domain && ing.Hostname != "" {
							newIngress = append(newIngress, zero_trust.TunnelCloudflaredConfigurationUpdateParamsConfigIngress{
								Hostname: cloudflare.F(ing.Hostname),
								Service:  cloudflare.F(ing.Service),
							})
						}
					}
					// Re-add the catch-all as the very last rule
					newIngress = append(newIngress, zero_trust.TunnelCloudflaredConfigurationUpdateParamsConfigIngress{
						Service: cloudflare.F("http_status:404"),
					})

					_, err = cf.ZeroTrust.Tunnels.Cloudflared.Configurations.Update(ctx, tunnelID, zero_trust.TunnelCloudflaredConfigurationUpdateParams{
						AccountID: cloudflare.F(tunnel.Spec.AccountID),
						Config:    cloudflare.F(zero_trust.TunnelCloudflaredConfigurationUpdateParamsConfig{Ingress: cloudflare.F(newIngress)}),
					})
					if err != nil {
						log.Error(err, "Failed to remove route during cleanup")
						return ctrl.Result{RequeueAfter: time.Minute * 1}, nil
					}
				}
			}

			controllerutil.RemoveFinalizer(tunnel, tunnelFinalizer)
			return ctrl.Result{}, r.Update(ctx, tunnel)
		}
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(tunnel, tunnelFinalizer) {
		log.Info("Adding finalizer")
		controllerutil.AddFinalizer(tunnel, tunnelFinalizer)
		return ctrl.Result{}, r.Update(ctx, tunnel)
	}

	log.Info("Listing Cloudflare tunnels", "AccountID", tunnel.Spec.AccountID)

	tunnels, err := cf.ZeroTrust.Tunnels.List(ctx, zero_trust.TunnelListParams{
		AccountID: cloudflare.F(tunnel.Spec.AccountID),
	})
	if err != nil {
		log.Error(err, "List tunnels error")
		return ctrl.Result{RequeueAfter: time.Minute * 1}, nil
	}

	var tunnelID string
	for _, t := range tunnels.Result {
		if t.Name == tunnel.Spec.SharedTunnelName {
			tunnelID = t.ID
			break
		}
	}
	if tunnelID == "" {
		log.Info("Tunnel not found", "Name", tunnel.Spec.SharedTunnelName)
		return ctrl.Result{RequeueAfter: time.Minute * 1}, nil
	}

	log.Info("Found tunnel ID", "TunnelID", tunnelID)

	config, err := cf.ZeroTrust.Tunnels.Cloudflared.Configurations.Get(ctx, tunnelID, zero_trust.TunnelCloudflaredConfigurationGetParams{
		AccountID: cloudflare.F(tunnel.Spec.AccountID),
	})
	if err != nil {
		log.Error(err, "Failed to get config")
		return ctrl.Result{RequeueAfter: time.Minute * 1}, nil
	}

	log.Info("Config retrieved", "TunnelID", tunnelID)

	// Check if already configured
	needsUpdate := false
	var updatedIngress []zero_trust.TunnelCloudflaredConfigurationUpdateParamsConfigIngress
	for _, ing := range config.Config.Ingress {
		if ing.Hostname == "" { // catch-all
			continue
		}
		if ing.Hostname == tunnel.Spec.Domain {
			// Check if service needs update
			expectedService := fmt.Sprintf("http://%s:%d", tunnel.Spec.ServiceName, tunnel.Spec.ServicePort)
			if ing.Service != expectedService {
				needsUpdate = true
				updatedIngress = append(updatedIngress, zero_trust.TunnelCloudflaredConfigurationUpdateParamsConfigIngress{
					Hostname: cloudflare.F(tunnel.Spec.Domain),
					Service:  cloudflare.F(expectedService),
				})
			} else {
				updatedIngress = append(updatedIngress, zero_trust.TunnelCloudflaredConfigurationUpdateParamsConfigIngress{
					Hostname: cloudflare.F(ing.Hostname),
					Service:  cloudflare.F(ing.Service),
				})
			}
		} else {
			updatedIngress = append(updatedIngress, zero_trust.TunnelCloudflaredConfigurationUpdateParamsConfigIngress{
				Hostname: cloudflare.F(ing.Hostname),
				Service:  cloudflare.F(ing.Service),
			})
		}
	}

	if !needsUpdate {
		// Check if domain is missing
		found := false
		for _, ing := range updatedIngress {
			if ing.Hostname == cloudflare.F(tunnel.Spec.Domain) {
				found = true
				break
			}
		}
		if !found {
			needsUpdate = true
			updatedIngress = append(updatedIngress, zero_trust.TunnelCloudflaredConfigurationUpdateParamsConfigIngress{
				Hostname: cloudflare.F(tunnel.Spec.Domain),
				Service:  cloudflare.F(fmt.Sprintf("http://%s:%d", tunnel.Spec.ServiceName, tunnel.Spec.ServicePort)),
			})
		}
	}

	if !needsUpdate {
		log.Info("Route already up to date, skipping update")
		return ctrl.Result{}, nil
	}

	// Add catch-all
	updatedIngress = append(updatedIngress, zero_trust.TunnelCloudflaredConfigurationUpdateParamsConfigIngress{
		Service: cloudflare.F("http_status:404"),
	})

	_, err = cf.ZeroTrust.Tunnels.Cloudflared.Configurations.Update(ctx, tunnelID, zero_trust.TunnelCloudflaredConfigurationUpdateParams{
		AccountID: cloudflare.F(tunnel.Spec.AccountID),
		Config:    cloudflare.F(zero_trust.TunnelCloudflaredConfigurationUpdateParamsConfig{Ingress: cloudflare.F(updatedIngress)}),
	})
	if err != nil {
		log.Error(err, "Failed to update config")
		return ctrl.Result{RequeueAfter: time.Minute * 1}, nil
	}

	return ctrl.Result{}, nil
}

func (r *TunnelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1alpha1.Tunnel{}).
		Complete(r)
}
