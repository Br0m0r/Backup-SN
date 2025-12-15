<template>
  <div class="profile-shell">
    <section class="profile-hero">
      <div class="hero-primary">
        <div class="avatar-container">
          <img :src="avatarUrl" alt="profile avatar" class="profile-avatar" />
          <button v-if="isOwnProfile" class="avatar-upload-btn" @click="triggerAvatarUpload" title="Change avatar">
            <span>📷</span>
          </button>
          <input 
            ref="avatarInput" 
            type="file" 
            accept="image/*" 
            @change="handleAvatarChange" 
            style="display: none"
          />
        </div>
        <div>
          <p v-if="displayHandle" class="nickname">{{ displayHandle }}</p>
          <h1>{{ displayName }}</h1>
          <p class="about">
            {{ aboutText }}
          </p>
          <p v-if="dateOfBirthDisplay && canViewFullProfile" class="dob-line">
            Born {{ dateOfBirthDisplay }}
          </p>
          <div v-if="canViewFullProfile" class="hero-stats">
            <div
              v-for="stat in stats"
              :key="stat.label"
              :class="['stat-block', { clickable: isPanelStat(stat.label) }]"
              @click="handleStatInteraction(stat.label)"
              @keydown.enter.prevent="handleStatInteraction(stat.label)"
              @keydown.space.prevent="handleStatInteraction(stat.label)"
              :role="isPanelStat(stat.label) ? 'button' : undefined"
              :tabindex="isPanelStat(stat.label) ? 0 : undefined"
            >
              <span>{{ stat.value }}</span>
              <small>{{ stat.label }}</small>
            </div>
          </div>
        </div>
      </div>
      <div class="hero-controls">
        <template v-if="isOwnProfile">
          <button class="privacy-toggle" @click="togglePrivacy">
            <span :class="['status-dot', isPrivate ? 'dot-private' : 'dot-public']"></span>
            {{ isPrivate ? 'Private mode' : 'Public mode' }}
          </button>
          <p class="privacy-copy">
          {{
            isPrivate
              ? 'Only accepted followers can see your content and profile fields.'
              : 'Everyone can view your highlights, posts, and profile fields.'
          }}
        </p>
        <div class="hero-buttons">
          <button class="ghost" @click="toggleInfoMode">{{ infoMode ? 'Hide info' : 'Info' }}</button>
        </div>
        </template>
        <template v-else>
          <div class="viewer-note">
            <p v-if="!canViewFullProfile" class="private-notice">
              🔒 This account is private. Follow to see their posts and profile.
            </p>
            <p v-else>You are viewing {{ displayName }}'s profile.</p>
            <button
              class="follow-btn"
              :disabled="followActionLoading || isFollowPending"
              @click="handleFollowToggle"
            >
              <span v-if="followActionLoading">...</span>
              <span v-else-if="isFollowPending">Request sent</span>
              <span v-else-if="isFollowing">Unfollow</span>
              <span v-else>Follow</span>
            </button>
          </div>
        </template>
      </div>
    </section>

    <Transition name="info-fade">
      <section v-if="showInfoGrid && canViewFullProfile" class="info-grid" :class="{ readonly: !isOwnProfile }">
        <article v-for="section in infoSections" :key="section.key">
          <header>
            <div>
              <small>{{ section.label }}</small>
              <h3>{{ displaySectionValue(section) }}</h3>
            </div>
            <div v-if="isOwnProfile" class="info-actions">
              <button class="ghost mini" @click="openEditor(section.key)">Edit</button>
            </div>
          </header>
        </article>
      </section>
    </Transition>

    <section v-if="canViewFullProfile" class="activity-panel">
      <h2 class="activity-title">Posts</h2>
      <div class="post-stack">
        <div v-if="postsLoading" class="loading">Loading posts...</div>
        <p v-else-if="postsError" class="empty-state">
          Failed to load posts. {{ postsError }}
        </p>
        <p v-else-if="!posts.length" class="empty-state">
          No posts yet.
        </p>
        <article
          v-else
          v-for="post in posts"
          :key="post.id"
          class="post-card"
          @click="navigateToPost(post.id)"
        >
          <header>
            <div class="author">
              <div class="avatar">{{ getInitials(post) }}</div>
              <div>
                <button
                  class="author-link"
                  type="button"
                  @click.stop="viewUserProfile(getPostOwnerId(post))"
                >
                  {{ displayName }}
                </button>
                <small>{{ formatTime(post.created_at) }} · {{ formatPrivacy(post.privacy_level) }}</small>
              </div>
            </div>
          </header>
          <h3 v-if="post.title" class="post-title">{{ post.title }}</h3>
          <p class="post-content">{{ post.content }}</p>
          <img v-if="post.image_path" :src="getImageUrl(post.image_path)" alt="Post image" class="post-image" />
        </article>
      </div>
    </section>

    <Transition name="side-panel-right">
      <aside v-if="isOwnProfile && infoMode && editingSection" class="side-panel editor-panel">
        <header>
          <div>
            <small>Edit field</small>
            <h3>{{ currentEdit?.label }}</h3>
          </div>
          <button class="icon-btn close-btn" aria-label="Close editor" @click="closeEditor">
            ✕
          </button>
        </header>
        <p class="panel-copy">
          Update your profile info and save. Changes sync to your account immediately.
        </p>
        <label class="panel-field">
          <span class="sr-only">{{ currentEdit?.label }}</span>
          <textarea
            v-if="currentEdit?.key === 'about'"
            v-model="draftValue"
            rows="6"
            class="editor-field"
            :placeholder="currentEdit?.label"
          />
          <input
            v-else
            v-model="draftValue"
            type="text"
            class="editor-field"
            :placeholder="currentEdit?.label"
          />
        </label>
        <p v-if="editorError" class="panel-error">{{ editorError }}</p>
        <div class="drawer-actions">
          <button class="ghost" @click="closeEditor">Cancel</button>
          <button class="cta" :disabled="savingEditor" @click="saveEditor">
            {{ savingEditor ? 'Saving...' : 'Save' }}
          </button>
        </div>
      </aside>
    </Transition>

    <Transition name="side-panel">
      <aside v-if="followersPanelOpen" class="side-panel followers-panel">
        <header>
          <div>
            <small>Followers</small>
            <h3>{{ followersTotalDisplay }} total</h3>
          </div>
          <button class="icon-btn close-btn" aria-label="Close followers list" @click="closeFollowersPanel">
            ✕
          </button>
        </header>
        <label class="panel-search">
          <span class="sr-only">Search followers</span>
          <input
            v-model="followerSearch"
            type="search"
            placeholder="Search followers..."
            autocomplete="off"
          />
        </label>
        <div class="panel-list">
          <div v-if="followersLoading" class="loading">Loading followers...</div>
          <p v-else-if="followersError" class="empty-state">
            Failed to load followers. {{ followersError }}
          </p>
          <template v-else>
            <article v-for="user in filteredFollowers" :key="user.id || user.handle" class="panel-row">
              <img :src="user.avatar" :alt="`${user.name} avatar`" />
              <div>
                <strong>{{ user.name }}</strong>
                <small>{{ user.handle }}</small>
              </div>
              <button class="ghost mini" @click="viewUserProfile(user.id)">View profile</button>
            </article>
            <p v-if="!filteredFollowers.length" class="empty-state">
              No followers match your search.
            </p>
          </template>
        </div>
      </aside>
    </Transition>

    <Transition name="side-panel">
      <aside v-if="followingPanelOpen" class="side-panel following-panel">
        <header>
          <div>
            <small>Following</small>
            <h3>{{ followingTotalDisplay }} accounts</h3>
          </div>
          <button class="icon-btn close-btn" aria-label="Close following list" @click="closeFollowingPanel">
            ✕
          </button>
        </header>
        <label class="panel-search">
          <span class="sr-only">Search following</span>
          <input
            v-model="followingSearch"
            type="search"
            placeholder="Search following..."
            autocomplete="off"
          />
        </label>
        <div class="panel-list">
          <div v-if="followingLoading" class="loading">Loading following...</div>
          <p v-else-if="followingError" class="empty-state">
            Failed to load following. {{ followingError }}
          </p>
          <template v-else>
            <article v-for="user in filteredFollowing" :key="user.id || user.handle" class="panel-row">
              <img :src="user.avatar" :alt="`${user.name} avatar`" />
              <div>
                <strong>{{ user.name }}</strong>
                <small>{{ user.handle }}</small>
              </div>
              <button class="ghost mini" @click="viewUserProfile(user.id)">View profile</button>
            </article>
            <p v-if="!filteredFollowing.length" class="empty-state">
              No accounts match your search.
            </p>
          </template>
        </div>
      </aside>
    </Transition>

    <Transition name="side-panel">
      <aside v-if="groupsPanelOpen" class="side-panel groups-panel">
        <header>
          <div>
            <small>Groups</small>
            <h3>{{ formatStat(groupsCount, '0') }} joined</h3>
          </div>
          <button class="icon-btn close-btn" aria-label="Close groups list" @click="closeGroupsPanel">
            ✕
          </button>
        </header>
        <label class="panel-search">
          <span class="sr-only">Search groups</span>
          <input
            v-model="groupSearch"
            type="search"
            placeholder="Search groups..."
            autocomplete="off"
          />
        </label>
        <div class="panel-list groups-list">
          <div v-if="groupsLoading" class="loading">Loading groups...</div>
          <p v-else-if="groupsError" class="empty-state">
            Failed to load groups. {{ groupsError }}
          </p>
          <template v-else>
            <article v-for="group in filteredGroups" :key="group.id || group.title" class="panel-row group-row">
              <div class="group-row__body">
                <strong>{{ group.title }}</strong>
                <small>{{ group.members }} members</small>
                <p>{{ group.desc }}</p>
              </div>
              <button class="ghost mini" @click="openGroup(group.id)">Open</button>
            </article>
            <p v-if="!filteredGroups.length" class="empty-state">
              No groups match your search.
            </p>
          </template>
        </div>
      </aside>
    </Transition>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { getToken, getUser } from '../stores/auth';
import { getFollowers, getFollowing, getUserProfile, updatePrivacy, updateProfile, uploadAvatar, followUser, unfollowUser, getFollowStatus } from '../services/usersService';
import { getUserPosts, getPostImageUrl } from '../services/postsService';
import { getMyGroups } from '../services/groupsService';
import { useToast } from '@/composables/useToast';
import { useAvatar } from '@/composables/useAvatar';
import { throttle } from '@/utils/timing';

const { success, error: showError } = useToast();
const { getUserAvatarUrl } = useAvatar();
const router = useRouter();
const props = defineProps({
  id: {
    type: [String, Number],
    default: null,
  },
});

const emit = defineEmits(['back']);

const isPrivate = ref(false);
const followerCount = ref(null);
const followingCount = ref(null);
const followers = ref([]);
const followersLoading = ref(false);
const followersError = ref('');
const profileUser = ref(null);
const viewerUserId = ref(null);
const following = ref([]);
const followingLoading = ref(false);
const followingError = ref('');
const posts = ref([]);
const postsLoading = ref(false);
const postsError = ref('');
const activeProfileId = ref(null);
const editingSection = ref('');
const infoMode = ref(false);
const draftValue = ref('');
const followersPanelOpen = ref(false);
const followingPanelOpen = ref(false);
const groupsPanelOpen = ref(false);
const followerSearch = ref('');
const followingSearch = ref('');
const groupSearch = ref('');
const editorError = ref('');
const savingEditor = ref(false);
const avatarInput = ref(null);
const uploadingAvatar = ref(false);
const myGroups = ref([]);
const groupsLoading = ref(false);
const groupsError = ref('');
const followStatus = ref('none');
const followActionLoading = ref(false);

const groupsCount = computed(() => myGroups.value.length);
const socialPanelsOpen = computed(() => followersPanelOpen.value || followingPanelOpen.value || groupsPanelOpen.value);

const stats = computed(() => [
  { label: 'Followers', value: formatStat(followerCount.value, '0') },
  { label: 'Following', value: formatStat(followingCount.value, '0') },
  { label: 'Groups', value: formatStat(groupsCount.value, '0') },
]);

const panelStatLabels = ['Followers', 'Following', 'Groups'];

const infoSections = ref([
  { key: 'firstName', label: 'First Name', value: 'Marina', visible: true },
  { key: 'lastName', label: 'Last Name', value: 'Pulse', visible: true },
  { key: 'dob', label: 'Date of Birth', value: '1994-08-17', visible: false },
  { key: 'nickname', label: 'Nickname', value: 'Neon Pilot', visible: true },
  { key: 'about', label: 'About Me', value: 'Collecting retro synths & designing social UX.', visible: true },
]);

const displayHandle = computed(() => {
  const nicknameSection = getSectionState('nickname');
  if (nicknameSection.visible) {
    const nicknameVal = nicknameSection.value || profileUser.value?.nickname || '';
    if (nicknameVal) {
      return nicknameVal.startsWith('@') ? nicknameVal : `@${nicknameVal}`;
    }

    const username = profileUser.value?.username;
    if (username) return `@${username}`;
  }

  return '';
});

function getSectionState(key) {
  return infoSections.value.find((section) => section.key === key) || { value: '', visible: false };
}

const displayName = computed(() => {
  const first = getSectionState('firstName');
  const last = getSectionState('lastName');
  const nicknameSection = getSectionState('nickname');
  const parts = [];

  if (first.visible && first.value) parts.push(first.value);
  if (last.visible && last.value) parts.push(last.value);

  if (parts.length) return parts.join(' ');
  if (nicknameSection.visible && nicknameSection.value) return nicknameSection.value;

  const user = profileUser.value;
  if (user) {
    const fallback = [user.first_name, user.last_name].filter(Boolean);
    if (fallback.length) return fallback.join(' ');
    if (user.nickname) return user.nickname;
    if (user.username) return user.username;
  }

  return 'Profile';
});

const avatarUrl = computed(() => {
  return getUserAvatarUrl(profileUser.value);
});

const aboutText = computed(() => {
  const aboutSection = getSectionState('about');
  if (aboutSection.visible && aboutSection.value) return aboutSection.value;
  return '';
});

const dateOfBirthDisplay = computed(() => {
  const dobSection = getSectionState('dob');
  if (!dobSection.visible || !dobSection.value) return '';
  return formatDateOfBirth(dobSection.value);
});

const isFollowing = computed(() => followStatus.value === 'accepted');
const isFollowPending = computed(() => followStatus.value === 'pending');
const isOtherProfile = computed(() => !isOwnProfile.value && !!activeProfileId.value);

// Allow viewing full profile if: own profile, public profile, or following a private profile
const canViewFullProfile = computed(() => {
  if (isOwnProfile.value) return true; // Own profile - show everything
  if (!isPrivate.value) return true; // Public profile - show everything
  return isFollowing.value; // Private profile - only show if following
});

const filteredFollowers = computed(() => {
  const term = followerSearch.value.trim().toLowerCase();
  const list = followers.value || [];
  if (!term) return list;
  return list.filter((user) => {
    return [user.name, user.handle].some((field) => field.toLowerCase().includes(term));
  });
});

const filteredFollowing = computed(() => {
  const term = followingSearch.value.trim().toLowerCase();
  const list = following.value || [];
  if (!term) return list;
  return list.filter((user) => {
    return [user.name, user.handle].some((field) => field.toLowerCase().includes(term));
  });
});

const filteredGroups = computed(() => {
  const term = groupSearch.value.trim().toLowerCase();
  const list = myGroups.value || [];
  if (!term) return list;
  return list.filter((group) => {
    return [group.title, group.desc].some((field) => (field || '').toLowerCase().includes(term));
  });
});

const currentEdit = computed(() => infoSections.value.find((section) => section.key === editingSection.value));
const followersTotalDisplay = computed(() =>
  formatStat(followerCount.value ?? followers.value?.length ?? 0, '0')
);
const followingTotalDisplay = computed(() =>
  formatStat(followingCount.value ?? following.value?.length ?? 0, '0')
);
const isOwnProfile = computed(() => {
  const viewerId = viewerUserId.value || getUser()?.id;
  const profileId = activeProfileId.value || (props.id ? Number(props.id) : viewerId);
  if (!viewerId || !profileId) return false;
  return String(viewerId) === String(profileId);
});
const showInfoGrid = computed(() => infoMode.value && isOwnProfile.value);

const togglePrivacy = throttle(async () => {
  if (!isOwnProfile.value) {
    showError('You can only update privacy on your own profile.');
    return;
  }

  const newPrivacy = !isPrivate.value;
  
  try {
    const token = getToken();
    if (!token) {
      console.error('No auth token available');
      return;
    }

    // Call backend API to update privacy
    await updatePrivacy(!newPrivacy, token); // Backend expects is_public (opposite of isPrivate)
    
    // Update local state on success
    isPrivate.value = newPrivacy;
    console.log(`Privacy updated: ${newPrivacy ? 'Private' : 'Public'}`);
  } catch (error) {
    console.error('Failed to update privacy:', error);
    alert('Failed to update privacy settings. Please try again.');
  }
}, 2000)

function formatCount(value) {
  if (value === null || value === undefined) return '0';
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(1).replace(/\.0$/, '')}M`;
  if (value >= 1_000) return `${(value / 1_000).toFixed(1).replace(/\.0$/, '')}K`;
  return String(value);
}

function formatStat(value, fallback) {
  if (value === null || value === undefined) return fallback;
  return formatCount(value);
}

function formatTime(timestamp) {
  if (!timestamp) return '';
  const date = new Date(timestamp);
  const now = new Date();
  const diff = Math.floor((now - date) / 1000);
  if (diff < 60) return `${diff}s ago`;
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  return `${Math.floor(diff / 86400)}d ago`;
}

function formatDateOfBirth(value) {
  if (!value) return '';
  const plain = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value);
  if (plain) {
    const [, y, m, d] = plain;
    return `${d}-${m}-${y}`;
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  const dd = String(date.getDate()).padStart(2, '0');
  const mm = String(date.getMonth() + 1).padStart(2, '0');
  const yyyy = date.getFullYear();
  return `${dd}-${mm}-${yyyy}`;
}

function displaySectionValue(section) {
  if (!section) return '';
  if (section.key === 'dob') {
    return formatDateOfBirth(section.value);
  }
  return section.value;
}

function normalizeGroup(group) {
  const title = group?.name || group?.title || 'Group';
  return {
    id: group?.id ?? title,
    title,
    desc: group?.description || 'No description provided.',
    members: group?.member_count ?? 0,
  };
}

function getImageUrl(path) {
  return getPostImageUrl(path);
}

function getPostOwnerId(post) {
  if (!post) return activeProfileId.value || profileUser.value?.id || null;
  return post.user_id ?? post.author?.id ?? activeProfileId.value ?? profileUser.value?.id ?? null;
}

function getInitials(post) {
  const user = profileUser.value;
  if (!user) return '??';
  const first = user.first_name?.[0] || '';
  const last = user.last_name?.[0] || '';
  return (first + last).toUpperCase() || user.username?.[0]?.toUpperCase() || '?';
}

function formatPrivacy(level) {
  if (level === 'public') return 'Public';
  if (level === 'private') return 'Private';
  if (level === 'almost_private') return 'Friends';
  return level;
}

function triggerAvatarUpload() {
  if (!isOwnProfile.value) return;
  avatarInput.value?.click();
}

async function handleAvatarChange(event) {
  if (!isOwnProfile.value) return;

  const file = event.target.files?.[0];
  if (!file) return;

  // Validate file type
  if (!file.type.startsWith('image/')) {
    showError('Please select an image file');
    return;
  }

  // Validate file size (max 10MB)
  if (file.size > 10 * 1024 * 1024) {
    showError('File size must be less than 10MB');
    return;
  }

  uploadingAvatar.value = true;
  try {
    const token = getToken();
    const result = await uploadAvatar(file, token);
    
    // Update profile user with new avatar path
    if (profileUser.value) {
      profileUser.value.avatar_path = result.avatar_path;
    }
    
    success('Avatar updated successfully!');
  } catch (error) {
    console.error('Failed to upload avatar:', error);
    showError(error.message || 'Failed to upload avatar');
  } finally {
    uploadingAvatar.value = false;
    // Reset input
    if (avatarInput.value) {
      avatarInput.value.value = '';
    }
  }
}

function applyUserProfile(user) {
  profileUser.value = user || null;

  infoSections.value = infoSections.value.map((section) => {
    switch (section.key) {
      case 'firstName':
        return { ...section, value: user?.first_name || '', visible: !!user?.first_name };
      case 'lastName':
        return { ...section, value: user?.last_name || '', visible: !!user?.last_name };
      case 'nickname':
        return { ...section, value: user?.nickname || '', visible: !!user?.nickname };
      case 'about':
        return { ...section, value: user?.about_me || '', visible: !!user?.about_me };
      case 'dob':
        // Show DOB if it exists and user has made it public (show_date_of_birth field from backend)
        // or default to showing it if the value exists
        const dobVisible = user?.show_date_of_birth !== undefined 
          ? user.show_date_of_birth 
          : !!user?.date_of_birth;
        return { ...section, value: user?.date_of_birth || '', visible: dobVisible };
      default:
        return section;
    }
  });
}

function updateSectionValue(key, value) {
  infoSections.value = infoSections.value.map((section) =>
    section.key === key ? { ...section, value } : section
  );
}

function getSectionValue(key) {
  const sectionVal = infoSections.value.find((section) => section.key === key)?.value;
  if (sectionVal !== undefined && sectionVal !== null) return sectionVal;

  // Fallback to profile user in case section state is out of sync
  const user = profileUser.value || {};
  switch (key) {
    case 'firstName':
      return user.first_name || '';
    case 'lastName':
      return user.last_name || '';
    case 'dob':
      return user.date_of_birth || '';
    case 'nickname':
      return user.nickname || '';
    case 'about':
      return user.about_me || '';
    default:
      return '';
  }
}

function shouldLoadFollowers() {
  // Allow loading followers for any profile
  return true;
}

function shouldLoadFollowing() {
  // Allow loading following for any profile
  return true;
}

function shouldLoadGroupData() {
  const currentUser = getUser();
  if (!currentUser?.id) return false;
  if (props.id === null || props.id === undefined) return true;
  return String(props.id) === String(currentUser.id);
}

function normalizeFollower(user) {
  const first = user.first_name || '';
  const last = user.last_name || '';
  const name = `${first} ${last}`.trim() || user.username || 'Follower';
  const handle = user.username ? `@${user.username}` : '@user';
  return {
    id: user.id ?? handle,
    name,
    handle,
    avatar: user.avatar_url || 'https://placehold.co/64x64/20223c/fff?text=FF',
  };
}

async function loadFollowing() {
  if (!shouldLoadFollowing()) {
    following.value = [];
    followingError.value = 'Unable to load following.';
    return;
  }

  const token = getToken();
  if (!token) {
    following.value = [];
    followingError.value = 'You need to be logged in to load following.';
    return;
  }

  // Use activeProfileId, props.id, or current user's ID
  const currentUserId = getUser()?.id;
  const targetId = activeProfileId.value || props.id || currentUserId;
  if (!targetId) {
    following.value = [];
    followingError.value = 'Unable to determine profile.';
    return;
  }

  followingLoading.value = true;
  followingError.value = '';

  try {
    const { following: list = [], count } = await getFollowing(token, targetId);
    following.value = list.map(normalizeFollower);
    if (count !== undefined && count !== null) {
      followingCount.value = count;
    } else {
      followingCount.value = following.value.length;
    }
  } catch (error) {
    console.error('Failed to load following:', error);
    followingError.value = error?.message || 'Unable to load following';
  } finally {
    followingLoading.value = false;
  }
}

async function loadGroups() {
  if (!shouldLoadGroupData()) {
    myGroups.value = [];
    groupsError.value = 'Groups list is available only for your profile.';
    return;
  }

  const token = getToken();
  if (!token) {
    myGroups.value = [];
    groupsError.value = 'You need to be logged in to load groups.';
    return;
  }

  groupsLoading.value = true;
  groupsError.value = '';

  try {
    const { groups = [] } = await getMyGroups(token);
    myGroups.value = groups.map(normalizeGroup);
  } catch (error) {
    console.error('Failed to load groups:', error);
    groupsError.value = error?.message || 'Unable to load groups';
    myGroups.value = [];
  } finally {
    groupsLoading.value = false;
  }
}

async function loadProfileStats() {
  const token = getToken();
  const viewerId = getUser()?.id;
  viewerUserId.value = viewerId || null;
  followStatus.value = 'none';

  const targetId = props.id ? Number(props.id) : viewerId;
  const parsedId = Number(targetId);

  if (!token || !parsedId || Number.isNaN(parsedId)) {
    console.warn('Cannot load profile stats without auth token and user id.');
    return;
  }

  try {
    const { profile } = await getUserProfile(parsedId, token);
    applyUserProfile(profile?.user);
    activeProfileId.value = profile?.user?.id || parsedId;
    followerCount.value = profile?.follower_count ?? 0;
    followingCount.value = profile?.following_count ?? 0;
    if (!isOwnProfile.value) {
      if (profile?.follow_status) {
        followStatus.value = profile.follow_status;
      } else if (profile?.is_following) {
        followStatus.value = 'accepted';
      } else {
        followStatus.value = 'none';
      }
    } else {
      followStatus.value = 'self';
    }

    if (profile?.user?.is_public_profile !== undefined) {
      isPrivate.value = !profile.user.is_public_profile;
    }

    await loadProfilePosts();
  } catch (error) {
    console.error('Failed to load profile stats:', error);
  }
}

async function loadProfilePosts() {
  const token = getToken();
  const userId = activeProfileId.value || props.id || getUser()?.id;
  const targetId = Number(userId);

  if (!token || !targetId || Number.isNaN(targetId)) {
    posts.value = [];
    postsError.value = 'Unable to load posts (missing user or token).';
    return;
  }

  postsLoading.value = true;
  postsError.value = '';

  try {
    const { posts: list = [] } = await getUserPosts(targetId, token);
    posts.value = list;
  } catch (error) {
    console.error('Failed to load profile posts:', error);
    postsError.value = error?.message || 'Unable to load posts';
    posts.value = [];
  } finally {
    postsLoading.value = false;
  }
}

async function loadFollowStatus() {
  const token = getToken();
  const targetId = activeProfileId.value || props.id;
  if (!token || !targetId || isOwnProfile.value) return;

  try {
    const { status } = await getFollowStatus(targetId, token);
    followStatus.value = status || 'none';
  } catch (error) {
    console.error('Failed to load follow status:', error);
  }
}

const handleFollowToggle = throttle(async () => {
  const token = getToken();
  const targetId = activeProfileId.value || props.id;
  if (!token || !targetId) {
    showError('You need to be logged in to follow users.');
    return;
  }
  if (isOwnProfile.value) return;

  followActionLoading.value = true;
  try {
    if (isFollowing.value) {
      await unfollowUser(targetId, token);
      followStatus.value = 'none';
      success('Unfollowed');
    } else {
      const result = await followUser(targetId, token);
      const nextStatus = result?.follow_status || (isPrivate.value ? 'pending' : 'accepted');
      followStatus.value = nextStatus;
      success(nextStatus === 'accepted' ? 'Now following' : 'Follow requested');
    }
  } catch (error) {
    console.error('Follow toggle failed:', error);
    showError(error?.message || 'Unable to update follow status');
  } finally {
    followActionLoading.value = false;
  }
}, 1000)

async function loadFollowers() {
  if (!shouldLoadFollowers()) {
    followers.value = [];
    followersError.value = 'Unable to load followers.';
    return;
  }

  const token = getToken();
  if (!token) {
    followers.value = [];
    followersError.value = 'You need to be logged in to load followers.';
    return;
  }

  // Use activeProfileId, props.id, or current user's ID
  const currentUserId = getUser()?.id;
  const targetId = activeProfileId.value || props.id || currentUserId;
  if (!targetId) {
    followers.value = [];
    followersError.value = 'Unable to determine profile.';
    return;
  }

  followersLoading.value = true;
  followersError.value = '';

  try {
    const { followers: list = [], count } = await getFollowers(token, targetId);
    followers.value = list.map(normalizeFollower);
    if (count !== undefined && count !== null) {
      followerCount.value = count;
    } else {
      followerCount.value = followers.value.length;
    }
  } catch (error) {
    console.error('Failed to load followers:', error);
    followersError.value = error?.message || 'Unable to load followers';
  } finally {
    followersLoading.value = false;
  }
}

function openEditor(key) {
  if (!infoMode.value || !isOwnProfile.value) return;
  // Toggle panel closed if clicking the same section while it's already open
  if (editingSection.value === key) {
    closeEditor();
    return;
  }
  editingSection.value = key;
  draftValue.value = getSectionValue(key);
  editorError.value = '';
}

function closeEditor() {
  editingSection.value = '';
  draftValue.value = '';
  editorError.value = '';
  savingEditor.value = false;
}

const saveEditor = throttle(async () => {
  if (!editingSection.value || savingEditor.value) return;

  const token = getToken();
  if (!token) {
    editorError.value = 'You must be logged in to update your profile.';
    return;
  }

  const value = draftValue.value;
  const trimmed = typeof value === 'string' ? value.trim() : value;
  const payload = {};

  switch (editingSection.value) {
    case 'firstName':
      payload.first_name = trimmed;
      break;
    case 'lastName':
      payload.last_name = trimmed;
      break;
    case 'dob':
      payload.date_of_birth = trimmed;
      break;
    case 'nickname':
      payload.nickname = trimmed;
      break;
    case 'about':
      payload.about_me = value;
      break;
    default:
      editorError.value = 'Unsupported field.';
      return;
  }

  savingEditor.value = true;
  editorError.value = '';

  try {
    const { user } = await updateProfile(payload, token);
    if (user) {
      applyUserProfile(user);
      // refresh draft to reflect saved value in case panel stays open after partial update
      draftValue.value = getSectionValue(editingSection.value);
      // also update the specific section immediately using the returned payload
      switch (editingSection.value) {
        case 'firstName':
          updateSectionValue('firstName', user.first_name || '');
          break;
        case 'lastName':
          updateSectionValue('lastName', user.last_name || '');
          break;
        case 'dob':
          updateSectionValue('dob', user.date_of_birth || '');
          break;
        case 'nickname':
          updateSectionValue('nickname', user.nickname || '');
          break;
        case 'about':
          updateSectionValue('about', user.about_me || '');
          break;
        default:
          break;
      }
      await loadProfileStats(); // pull latest profile data from backend (counts/visibility/etc.)
    } else {
      editorError.value = 'Update failed: no user returned.';
      savingEditor.value = false;
      return;
    }
    closeEditor();
  } catch (error) {
    console.error('Failed to update profile field:', error);
    editorError.value = error?.response?.data?.error || error?.message || 'Unable to update profile right now.';
    savingEditor.value = false;
  }
}, 1000)

watch(
  () => editingSection.value,
  (key) => {
    if (!key) return;
    draftValue.value = getSectionValue(key);
    editorError.value = '';
  }
);

function toggleInfoMode() {
  if (!isOwnProfile.value) return;
  infoMode.value = !infoMode.value;
  if (!infoMode.value) {
    closeEditor();
  }
}

function isPanelStat(label) {
  // Allow Followers and Following for all profiles, Groups only for own profile
  if (label === 'Followers' || label === 'Following') {
    return panelStatLabels.includes(label);
  }
  return isOwnProfile.value && panelStatLabels.includes(label);
}

function handleStatInteraction(label) {
  if (!isPanelStat(label)) return;
  if (label === 'Followers') {
    followersPanelOpen.value ? closeFollowersPanel() : openFollowersPanel();
  } else if (label === 'Following') {
    followingPanelOpen.value ? closeFollowingPanel() : openFollowingPanel();
  } else if (label === 'Groups') {
    groupsPanelOpen.value ? closeGroupsPanel() : openGroupsPanel();
  }
}

function closeAllPanels() {
  closeFollowersPanel();
  closeFollowingPanel();
  closeGroupsPanel();
}

function openFollowersPanel() {
  closeAllPanels();
  followersPanelOpen.value = true;
  loadFollowers();
}

function openFollowingPanel() {
  closeAllPanels();
  followingPanelOpen.value = true;
  loadFollowing();
}

async function openGroupsPanel() {
  closeAllPanels();
  groupsPanelOpen.value = true;
}

function closeFollowersPanel() {
  followersPanelOpen.value = false;
  followerSearch.value = '';
}

function closeFollowingPanel() {
  followingPanelOpen.value = false;
  followingSearch.value = '';
}

function closeGroupsPanel() {
  groupsPanelOpen.value = false;
  groupSearch.value = '';
}

function navigateToPost(postId) {
  if (!postId) return;
  router.push(`/post/${postId}`);
}

const handleSidePanelClickOutside = (event) => {
  if (!socialPanelsOpen.value) return;

  const target = event.target;
  if (!(target instanceof Element)) return;
  if (target.closest('.side-panel')) return;
  if (target.closest('.hero-stats')) return;

  closeAllPanels();
};

watch(socialPanelsOpen, (isOpen) => {
  if (isOpen) {
    document.addEventListener('click', handleSidePanelClickOutside);
  } else {
    document.removeEventListener('click', handleSidePanelClickOutside);
  }
});

function viewUserProfile(userId) {
  if (!userId) {
    showError('Unable to view profile');
    return;
  }
  // Close the panel
  closeFollowersPanel();
  closeFollowingPanel();
  // Navigate to the user's profile
  router.push(`/profile/${userId}`);
}

function openGroup(groupId) {
  if (!groupId) {
    showError('Unable to open group');
    return;
  }
  // Close the panel
  closeGroupsPanel();
  // Navigate to the group page
  router.push(`/groups/${groupId}`);
}

onBeforeUnmount(() => {
  document.removeEventListener('click', handleSidePanelClickOutside);
});

onMounted(async () => {
  if (!props.id) {
    applyUserProfile(getUser());
  }
  await loadProfileStats();
  await loadFollowStatus();
  loadFollowers();
  loadFollowing();
  await loadGroups();
});

watch(
  () => props.id,
  async () => {
    applyUserProfile(null);
    followStatus.value = 'none';
    followActionLoading.value = false;
    closeEditor();
    infoMode.value = false;
    await loadProfileStats();
    await loadFollowStatus();
    loadFollowers();
    loadFollowing();
    loadProfilePosts();
    await loadGroups();
    closeAllPanels();
  }
);
</script>

<style scoped>
.profile-shell {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2rem;
  max-width: 1400px;
  margin: 0 auto;
  width: 100%;
}

.profile-hero {
  border-radius: 1.75rem;
  padding: clamp(1.5rem, 4vw, 2.5rem);
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: linear-gradient(120deg, rgba(0, 247, 255, 0.1), rgba(255, 0, 230, 0.08));
  display: flex;
  flex-wrap: wrap;
  gap: 1.5rem;
  justify-content: space-between;
  width: 100%;
  max-width: 1100px;
}

.hero-primary {
  display: flex;
  gap: 1.5rem;
  align-items: center;
  min-width: 260px;
}

.avatar-container {
  position: relative;
  width: 140px;
  height: 140px;
}

.profile-avatar {
  width: 100%;
  height: 100%;
  border-radius: 1.5rem;
  border: 2px solid rgba(255, 255, 255, 0.15);
  object-fit: cover;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.45);
}

.avatar-upload-btn {
  position: absolute;
  bottom: 8px;
  right: 8px;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  border: 2px solid rgba(255, 255, 255, 0.2);
  background: rgba(5, 6, 13, 0.9);
  color: var(--neon-cyan);
  font-size: 1.2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s ease;
  backdrop-filter: blur(10px);
}

.avatar-upload-btn:hover {
  background: rgba(0, 247, 255, 0.15);
  border-color: var(--neon-cyan);
  box-shadow: 0 0 15px rgba(0, 247, 255, 0.4);
  transform: scale(1.05);
}

.hero-primary h1 {
  margin: 0.35rem 0;
  font-size: clamp(1.8rem, 4vw, 2.4rem);
}

.nickname {
  text-transform: uppercase;
  letter-spacing: 0.2em;
  font-size: 0.7rem;
  color: var(--neon-cyan);
  margin: 0;
}

.about {
  color: var(--text-muted);
  max-width: 36ch;
}

.dob-line {
  color: var(--text-muted);
  margin: 0.2rem 0 0.4rem 0;
  font-size: 0.95rem;
}

.hero-stats {
  display: flex;
  gap: 1.5rem;
  margin-top: 1rem;
}

.stat-block {
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
  transition: color 0.2s ease;
}

.stat-block.clickable {
  cursor: pointer;
}

.stat-block.clickable:hover span,
.stat-block.clickable:focus-visible span {
  color: var(--neon-cyan);
}

.hero-stats span {
  font-size: 1.3rem;
  font-weight: 600;
}

.hero-stats small {
  color: var(--text-muted);
  letter-spacing: 0.05em;
}

.hero-controls {
  max-width: 320px;
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}

.viewer-note {
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
  color: var(--text-muted);
  max-width: 32ch;
}

.private-notice {
  color: var(--neon-pink);
  font-weight: 500;
  text-align: center;
}

.follow-btn {
  border: 1px solid rgba(255, 255, 255, 0.2);
  background: linear-gradient(120deg, var(--neon-cyan), var(--neon-pink));
  color: #05060d;
  border-radius: 999px;
  padding: 0.55rem 1.3rem;
  font-weight: 700;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, opacity 0.2s ease;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.25);
}

.follow-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.follow-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 14px 30px rgba(0, 0, 0, 0.3);
}

.privacy-toggle {
  display: inline-flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.65rem 1rem;
  border-radius: 999px;
  border: 1px solid var(--border-glow);
  background: rgba(3, 5, 12, 0.6);
  color: inherit;
  cursor: pointer;
}

.status-dot {
  width: 0.75rem;
  height: 0.75rem;
  border-radius: 999px;
  box-shadow: 0 0 10px currentColor;
}

.dot-public {
  color: var(--neon-cyan);
  background: var(--neon-cyan);
}

.dot-private {
  color: var(--neon-pink);
  background: var(--neon-pink);
}

.hero-buttons {
  display: flex;
  gap: 0.75rem;
}

.cta {
  border: none;
  border-radius: 999px;
  padding: 0.6rem 1.4rem;
  background: linear-gradient(120deg, var(--neon-cyan), var(--neon-pink));
  color: #05060d;
  font-weight: 600;
  cursor: pointer;
  box-shadow: 0 10px 25px rgba(255, 0, 230, 0.3);
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
   width: 100%;
  max-width: 1100px;
}

.info-grid.readonly .info-actions {
  display: none;
}

.info-fade-enter-active,
.info-fade-leave-active {
  transition: opacity 0.35s ease, transform 0.35s ease;
}

.info-fade-enter-from,
.info-fade-leave-to {
  opacity: 0;
  transform: translateY(16px) scale(0.98);
}

.info-grid article {
  padding: 1.25rem;
  border-radius: 1.25rem;
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(6, 8, 18, 0.8);
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.info-grid h3 {
  margin: 0.35rem 0 0;
}

.info-actions {
  display: flex;
  gap: 0.35rem;
  align-items: center;
}

.icon-btn {
  width: 2rem;
  height: 2rem;
  border-radius: 0.75rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(8, 10, 22, 0.7);
  display: grid;
  place-items: center;
  cursor: pointer;
}

.ghost {
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: inherit;
  border-radius: 999px;
  padding: 0.45rem 1.1rem;
  cursor: pointer;
}

.ghost.mini {
  padding: 0.3rem 0.9rem;
}

.activity-panel {
  border-radius: 1.5rem;
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(5, 6, 13, 0.85);
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  width: 100%;
  max-width: 900px;
}

.activity-title {
  margin: 0;
  font-size: 1.4rem;
  color: var(--neon-cyan);
}

.post-stack {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.post-card {
  padding: 1.5rem;
  border-radius: 1.25rem;
  border: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(8, 10, 24, 0.8);
  box-shadow: 0 12px 25px rgba(0, 0, 0, 0.35);
  backdrop-filter: blur(12px);
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
}

.post-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 16px 30px rgba(0, 0, 0, 0.4);
}

.post-card header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.85rem;
}

.author {
  display: flex;
  gap: 0.85rem;
  align-items: center;
}

.author-link {
  background: transparent;
  border: none;
  padding: 0;
  margin: 0;
  color: inherit;
  font: inherit;
  font-weight: 700;
  cursor: pointer;
  text-align: left;
}

.author-link:hover {
  color: var(--neon-cyan);
}

.author .avatar {
  width: 2.75rem;
  height: 2.75rem;
  border-radius: 999px;
  border: 2px solid rgba(255, 255, 255, 0.08);
  background: linear-gradient(135deg, var(--neon-cyan), var(--neon-pink));
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 0.9rem;
  color: #05060d;
}

.author small {
  color: var(--text-muted);
}

.post-title {
  font-size: 1.1rem;
  font-weight: 600;
  margin: 0 0 0.5rem 0;
  color: var(--neon-cyan);
}

.post-content {
  line-height: 1.6;
  margin: 0 0 1rem 0;
  word-wrap: break-word;
  word-break: break-word;
  overflow-wrap: break-word;
  white-space: pre-wrap;
  max-width: 100%;
  overflow: hidden;
}

.post-image {
  width: 100%;
  max-height: 500px;
  object-fit: cover;
  border-radius: 0.75rem;
  margin-top: 1rem;
}

.editor-panel {
  position: fixed;
  top: 0;
  bottom: 0;
  right: 0;
  left: auto;
  width: min(380px, 82vw);
  border-right: none;
  border-left: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(5, 6, 13, 0.92);
  box-shadow: -10px 0 30px rgba(0, 0, 0, 0.35);
  z-index: 31;
}

.panel-copy {
  color: var(--text-muted);
  margin: 0;
}

.panel-field {
  width: 100%;
}

.panel-field .editor-field {
  width: 100%;
  padding: 0.65rem 0.85rem;
  border-radius: 0.9rem;
  border: 1px solid rgba(255, 255, 255, 0.2);
  background: rgba(8, 9, 20, 0.8);
  color: inherit;
}

.panel-field textarea.editor-field {
  min-height: 160px;
  resize: vertical;
}

.panel-error {
  margin: 0;
  color: var(--neon-pink);
  font-size: 0.9rem;
}

.drawer-actions {
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
  margin-top: 1rem;
}

.side-panel {
  position: fixed;
  top: 0;
  bottom: 0;
  left: 0;
  width: min(360px, 85vw);
  padding: 1.75rem 1.5rem;
  background: rgba(5, 6, 13, 0.92);
  border-right: 1px solid rgba(255, 255, 255, 0.08);
  box-shadow: 10px 0 30px rgba(0, 0, 0, 0.35);
  backdrop-filter: blur(18px);
  display: flex;
  flex-direction: column;
  gap: 1rem;
  z-index: 30;
}

.side-panel header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.side-panel header small {
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--text-muted);
}

.close-btn {
  border-radius: 999px;
  width: 2.25rem;
  height: 2.25rem;
  border: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(0, 0, 0, 0.35);
  color: rgba(255, 255, 255, 0.85);
  font-size: 1.1rem;
  transition: border-color 0.2s ease, color 0.2s ease, background 0.2s ease;
}

.close-btn:hover,
.close-btn:focus-visible {
  border-color: rgba(0, 247, 255, 0.6);
  color: var(--neon-cyan);
  background: rgba(0, 0, 0, 0.55);
}

.panel-search {
  width: 100%;
}

.panel-search input {
  width: 100%;
  padding: 0.65rem 0.85rem;
  border-radius: 0.9rem;
  border: 1px solid rgba(255, 255, 255, 0.2);
  background: rgba(8, 9, 20, 0.8);
  color: inherit;
}

.panel-list {
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.panel-row {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  padding: 0.9rem;
  border-radius: 1rem;
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(7, 8, 18, 0.85);
}

.editor-panel .panel-field {
  width: 100%;
}

.editor-panel .editor-field:focus {
  outline: none;
  border-color: var(--neon-cyan);
  box-shadow: 0 0 0 3px rgba(0, 247, 255, 0.15);
}

.panel-row img {
  width: 44px;
  height: 44px;
  border-radius: 50%;
}

.groups-list .panel-row {
  align-items: flex-start;
}

.group-row__body {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.group-row__body p {
  margin: 0.1rem 0 0;
  color: var(--text-muted);
  font-size: 0.85rem;
}

.events-list .panel-row {
  flex-direction: column;
  align-items: flex-start;
  gap: 0.4rem;
}

.event-row__meta {
  display: flex;
  justify-content: space-between;
  width: 100%;
  gap: 0.5rem;
}

.event-row__meta small {
  color: var(--text-muted);
}

.event-row p {
  margin: 0;
  color: var(--text-muted);
}

.event-location {
  font-size: 0.85rem;
  color: var(--neon-cyan);
}

.empty-state {
  text-align: center;
  color: var(--text-muted);
  padding: 1rem 0;
}

.loading {
  color: var(--text-muted);
  padding: 0.75rem 0;
}

.side-panel-enter-active,
.side-panel-leave-active {
  transition: transform 0.35s ease, opacity 0.3s ease;
}

.side-panel-enter-from,
.side-panel-leave-to {
  transform: translateX(-100%);
  opacity: 0;
}

.side-panel-right-enter-active,
.side-panel-right-leave-active {
  transition: transform 0.35s ease, opacity 0.3s ease;
}

.side-panel-right-enter-from,
.side-panel-right-leave-to {
  transform: translateX(100%);
  opacity: 0;
}

/* Ensure editor panel anchors to the right, overriding default side-panel left positioning */
.editor-panel {
  left: auto !important;
  right: 0 !important;
  border-left: 1px solid rgba(255, 255, 255, 0.08);
  border-right: none;
  box-shadow: -10px 0 30px rgba(0, 0, 0, 0.35);
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  border: 0;
}

@media (max-width: 768px) {
  .hero-primary {
    flex-direction: column;
    align-items: flex-start;
  }

  .hero-stats {
    flex-wrap: wrap;
  }

  .hero-controls {
    width: 100%;
  }
}
</style>
