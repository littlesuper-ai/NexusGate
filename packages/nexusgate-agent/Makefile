# OpenWrt package Makefile for nexusgate-agent
include $(TOPDIR)/rules.mk

PKG_NAME:=nexusgate-agent
PKG_VERSION:=1.0.0
PKG_RELEASE:=1

include $(INCLUDE_DIR)/package.mk

define Package/nexusgate-agent
  SECTION:=utils
  CATEGORY:=Utilities
  TITLE:=NexusGate device agent
  DEPENDS:=+curl +mosquitto-client-ssl +jq
  PKGARCH:=all
endef

define Package/nexusgate-agent/description
  Agent for NexusGate enterprise router management platform.
  Handles device registration, heartbeat reporting, config sync, and remote commands.
endef

define Package/nexusgate-agent/install
	$(INSTALL_DIR) $(1)/usr/bin
	$(INSTALL_BIN) ./files/nexusgate-agent.sh $(1)/usr/bin/nexusgate-agent

	$(INSTALL_DIR) $(1)/etc/config
	$(INSTALL_CONF) ./files/nexusgate.conf $(1)/etc/config/nexusgate

	$(INSTALL_DIR) $(1)/etc/init.d
	$(INSTALL_BIN) ./files/nexusgate-agent.init $(1)/etc/init.d/nexusgate-agent
endef

$(eval $(call BuildPackage,nexusgate-agent))
